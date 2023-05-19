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

var requestUrlList []string

func getRecommendations(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Received Request for list Recommendations")

	var getRecommendations RecommendationInput
	var recommendations []ListRecommendations
	var rec *RecommendationItem
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

		rec = getById(id)

	}
	//if recommendationid is missing but clustername and namespace provided:
	if namespace != "" && clusterName != "" && recommendationId == "" {

		getByClusterNamespace(clusterName, namespace)

		// if missing both namespace and recommendation or if missing both namespace and clustername error
	} else if namespace == "" && recommendationId == "" || namespace == "" && clusterName == "" {
		klog.V(5).Infof("Request missing both RecommendationId and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both  RecommendationId and Namespace cannot be empty.\"}", http.StatusBadRequest)
		return
	}

	client := utils.HTTPClient()

	for _, requests := range requestUrlList {

		req, err := http.NewRequest("GET", requests, nil)

		if ok := helpers.ErrorHandlingRequests(w, err); !ok {
			rec.RecommendationStatus = fmt.Sprint("Error", err)
			return
		}

		res, err := client.Do(req)
		if err != nil {
			rec.RecommendationStatus = fmt.Sprint("Error", err)

			klog.Errorf("Error when calling listRecommendations %v", err)

		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			rec.RecommendationStatus = fmt.Sprint("Error", err)

			klog.Errorf("Error reading data from the response body %v", err)

		}

		if err := json.Unmarshal(body, &recommendations); err != nil {
			rec.RecommendationStatus = fmt.Sprint("Error", err)

			klog.Errorf("Cannot unmarshal response data: %v", err)

		}

		_, err = w.Write([]byte(body))

		if err != nil {
			rec.RecommendationStatus = fmt.Sprint("Error", err)

			klog.Warning("Unexpected error processing the response. ", err.Error())
			http.Error(w, "Unexpected error processing the response", http.StatusInternalServerError)
			return
		}

		rec.RecommendationStatus = fmt.Sprint("Recieved Recommendation")

		klog.V(4).Info("Received recommendations")

	}
}

func getById(id string) *RecommendationItem {

	var recitem *RecommendationItem
	for _, rec := range Recommendationstore.data {
		if rec.RecommendationID == id {
			recitem = rec
			for dep, deplist := range rec.Recommendation {
				for _, conlist := range deplist {
					for con := range conlist {

						containerRequestUrl := fmt.Sprint(list_recommendations_url + "?" + "experiment_name=" + "ns_" + rec.Cluster + "_" + rec.Namespace + "_" + id + "-" + dep + "-" + con)
						requestUrlList = append(requestUrlList, containerRequestUrl)

					}
				}
			}
		}
	}
	return recitem
}

func getByClusterNamespace(cluster string, namespace string) []string {

	for _, rec := range Recommendationstore.data {
		if rec.Cluster == cluster && rec.Namespace == namespace {
			for dep, deplist := range rec.Recommendation {
				for _, conlist := range deplist {
					for con := range conlist {

						containerRequestUrl := fmt.Sprint(list_recommendations_url + "?" + "experiment_name=" + "ns_" + rec.Cluster + "_" + rec.Namespace + "_" + rec.RecommendationID + "-" + dep + "-" + con)
						requestUrlList = append(requestUrlList, containerRequestUrl)

					}
				}
			}
		}
	}
	return requestUrlList
}
