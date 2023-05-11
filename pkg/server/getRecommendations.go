package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/golang/gddo/httputil/header"
	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/utils"

	klog "k8s.io/klog/v2"
)

var list_recommendations_url = config.Cfg.KruizeURL + "/listRecommendations"

//struct to capture input for getRecommendations request
type RecommendationInput []struct {
	RecommendationId string `json:"recommendation_id"`
	Namespace        string `json:"namespace"`
	ClusterName      string `json:"cluster_name"`
}

func getRecommendations(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Received Request for list Recommendations")

	var getRecommendations RecommendationInput
	var requestUrlList []string
	var recommendations []ListRecommendations
	var RecommendationStatusGlobal = RecommendationStatusMap{RecommendationStatus: make(map[string]string)}

	//check content type is json
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	//decode user's input:
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&getRecommendations)

	if ok := helpers.ErrorHandlingRequests(w, err); !ok {
		return
	}

	recommendationId := getRecommendations[0].RecommendationId
	namespace := getRecommendations[0].Namespace
	clusterName := getRecommendations[0].ClusterName

	//if recommendationid provided default to use recommendation id
	if recommendationId != "" {
		reqParts := strings.Split(recommendationId, "_")
		id := reqParts[3]

		for _, RecommendationIDMap := range RecommendationIDMaps {
			rec := RecommendationIDMap.RecommendationID[id]
			containerRequestUrl := fmt.Sprint(list_recommendations_url + "?" + "experiment_name=" + rec)
			requestUrlList = append(requestUrlList, containerRequestUrl)
		}
	}
	//if recommendationid is missing but clustername and namespace provided:
	if namespace != "" && clusterName != "" && recommendationId == "" {
		var id string
		clusterNamespace := clusterName + "_" + namespace

		id = ClusterNamespaceMaps[0].ClusterNamespace[clusterNamespace]
		for _, RecommendationIDMap := range RecommendationIDMaps {
			rec := RecommendationIDMap.RecommendationID[id]
			containerRequestUrl := fmt.Sprint(list_recommendations_url + "?" + "experiment_name=" + rec)
			requestUrlList = append(requestUrlList, containerRequestUrl)

		}
		// if missing both namespace and recommendation or if missing both namespace and clustername error
	} else if namespace == "" && recommendationId == "" || namespace == "" && clusterName == "" {
		klog.V(5).Infof("Request missing both RecommendationId and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both  RecommendationId and Namespace cannot be empty.\"}", http.StatusBadRequest)
		return
	}

	// example of request: http://<ip>:<kruize port>/listRecommendations?experiment_name=
	// ns_local-cluster_open-cluster-management-observability_00465750-observability-alertmanager-config-reloader

	// make listRecommendations requests to Kruize:
	client := utils.HTTPClient()
	badStatus := "Error"

	for _, requests := range requestUrlList {
		req, err := http.NewRequest("GET", requests, nil)

		if ok := helpers.ErrorHandlingRequests(w, err); !ok {
			return
		}

		res, err := client.Do(req)
		if err != nil {
			klog.Errorf("Error when calling listRecommendations %v", err)
			RecommendationStatusGlobal.RecommendationStatus[requests] = badStatus

		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			klog.Errorf("Error reading data from the response body %v", err)
			RecommendationStatusGlobal.RecommendationStatus[requests] = badStatus

		}

		if err := json.Unmarshal(body, &recommendations); err != nil {
			klog.Errorf("Cannot unmarshal response data: %v", err)
			RecommendationStatusGlobal.RecommendationStatus[requests] = badStatus

		}

		_, err = w.Write([]byte(body))
		if err != nil {
			klog.Warning("Unexpected error processing the response. ", err.Error())
			http.Error(w, "Unexpected error processing the response", http.StatusInternalServerError)
			return
		}
		klog.V(4).Info("Received recommendations")
		status := "Received recommendations"

		RecommendationStatusGlobal.RecommendationStatus[requests] = status

	}
}
