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
type GetRecommendations []struct {
	RecommendationId string `json:"recommendation_id"`
	Namespace        string `json:"namespace"`
	ClusterName      string `json:"cluster_name"`
}

var GetRecommendationsQueue chan GetRecommendations

func init() {
	GetRecommendationsQueue = make(chan GetRecommendations)
}

func getRecommendations(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Received Request for list Recommendations")

	var getRecommendations GetRecommendations

	// context := r.Context()

	//check content type is json
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&getRecommendations)

	if ok := helpers.ErrorHandlingRequests(w, err); !ok {
		return
	}

	recommendationId := getRecommendations[0].RecommendationId
	namespace := getRecommendations[0].Namespace
	clusterName := getRecommendations[0].ClusterName

	var requestUrlList []string

	//if there is a namespace provided and recommendation id default to use recommendation id
	if namespace == "" && recommendationId != "" || namespace != "" && recommendationId != "" {

		reqParts := strings.Split(recommendationId, "_")
		clusterNamespace := fmt.Sprintf("%s_%s", reqParts[1], reqParts[2])
		id := fmt.Sprint(strings.Split(recommendationId, "_")[3])

		for deploymentName, deployment := range NcID.NamespaceClusters[clusterNamespace].Recommendations[id].Deployments {
			for _, containerStatus := range deployment.ContainerStatuses {
				for containerName, status := range containerStatus {
					klog.V(5).Infof("Container name: %s, Status: %s\n", containerName, status)

					containerRequestUrl := fmt.Sprint(list_recommendations_url + "?" + "experiment_name=" + "ns_" + clusterNamespace + "_" + recommendationId + "-" + deploymentName + "-" + containerName)
					requestUrlList = append(requestUrlList, containerRequestUrl)
				}
			}
		}
	}
	if namespace != "" && clusterName != "" && recommendationId == "" {
		clusterNamespace := clusterName + "_" + namespace
		nc := NcID.NamespaceClusters[clusterNamespace]
		for id, recommendation := range nc.Recommendations {
			for deploymentName, deployment := range recommendation.Deployments {
				for _, containerStatus := range deployment.ContainerStatuses {
					for containerName, status := range containerStatus {
						klog.V(5).Infof("Container name: %s, Status: %s\n", containerName, status)

						containerRequestUrl := fmt.Sprint(list_recommendations_url + "?" + "experiment_name=" + "ns_" + clusterNamespace + "_" + id + "-" + deploymentName + "-" + containerName)

						requestUrlList = append(requestUrlList, containerRequestUrl)
					}
				}
			}
		}
		// if missing both namespace and recommendation or if missing both namespace and clustername error
	} else if namespace == "" && recommendationId == "" || namespace == "" && clusterName == "" {
		klog.V(5).Infof("Request missing both RecommendationId and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both  RecommendationId and Namespace cannot be empty.\"}", http.StatusBadRequest)
		return
	}

	// ex: http://<ip>:<kruize port>/listRecommendations?experiment_name=
	// ns_local-cluster_open-cluster-management-observability_00465750-observability-alertmanager-config-reloader

	client := utils.HTTPClient()
	var recommendations []ListRecommendations

	// request per container:
	for _, requests := range requestUrlList {
		req, err := http.NewRequest("GET", requests, nil)

		if ok := helpers.ErrorHandlingRequests(w, err); !ok {
			return
		}

		res, err := client.Do(req)
		if err != nil {
			klog.Errorf("Error when calling listRecommendations %v", err)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			klog.Errorf("Error reading data from the response body %v", err)
		}

		if err := json.Unmarshal(body, &recommendations); err != nil {
			klog.Errorf("Cannot unmarshal response data: %v", err)

		}
	}
}
