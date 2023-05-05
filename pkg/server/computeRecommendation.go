package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/prometheus"
	klog "k8s.io/klog/v2"
)

var CreateQueue chan Request

func init() {
	CreateQueue = make(chan Request)
}

type recommendation []struct {
	ClusterName         string `json:"clusterName"`
	Namespace           string `json:"namespace"`
	Application         string `json:"application"`
	MeasurementDuration string `json:"measurement_duration"` //ex: "15min"
}

type Request struct {
	RequestName    string
	Workloads      map[string][]string
	RequestContext context.Context
}

type Deployment struct {
	ContainerStatuses []map[string]string `json:"containerStatuses"`
	// 	//ex. [{container1 : status1}, {container2: status2}]
}

type Recommendation struct {
	Deployments map[string]Deployment `json:"containerStatuses"`
	//ex. {deployment-name: Deployment
}

type NamespaceCluster struct {
	Recommendations map[string]Recommendation `json:"recommendations"`
	//ex. {00737189 : Recommendation}
}

type NamespaceClusterID struct {
	NamespaceClusters map[string]NamespaceCluster `json:"namespace_clusterid"`
	//ex. {local-cluster_open-cluster-management: NamespaceCluster}
}

//ex:
// {local-cluster_open-cluster-management:
// {
//	{00737189 :
// 		{deployment-name:
// 		[
// 			{container1: status1},
// 			{container2: status2}, ...
// 		]
// 	}
// }
// }
//

// API implementation for /computeRecommendations
func computeRecommendations(w http.ResponseWriter, r *http.Request) {
	klog.V(5).Infof("Received Request for compute Recommendations")
	var newRecommendation recommendation
	requestIdMap := make(map[string]string) //ex: clustername-namespace:"id-12345"
	var requestName string

	//create context from request
	context := r.Context()

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
	err := dec.Decode(&newRecommendation)

	//error handling for decoding request body:
	if ok := helpers.ErrorHandlingRequests(w, err); !ok {
		return
	}

	clusterName := newRecommendation[0].ClusterName
	nameSpace := newRecommendation[0].Namespace
	appName := newRecommendation[0].Application
	//get the clusterID (cluster name with namespace or applicaiton):
	//use namespace:
	if newRecommendation[0].Application == "" && nameSpace != "" {
		requestName = fmt.Sprintf("ns_%s_%s", clusterName, nameSpace)
	}
	//use application , not supported on Dev Preview
	if newRecommendation[0].Application != "" && newRecommendation[0].Namespace == "" {
		requestName = fmt.Sprintf("app_%s_%s", clusterName, appName)
		// if both applications and namespace is empty return
	} else if newRecommendation[0].Application == "" && newRecommendation[0].Namespace == "" {
		klog.V(4).Info("Request missing both Application and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both Application and Namespace cannot be empty.\"}", http.StatusBadRequest)
		return
	}

	//if id for clusterName already exists then we don't need to generate new one
	if _, found := requestIdMap[requestName]; !found {
		uid := helpers.GenerateID()
		requestIdMap[requestName] = uid
		requestName = requestName + "_" + uid
	}

	//get the deployments and containers:
	deployments, err := prometheus.GetLabels(clusterName, nameSpace)

	//createExperiment with data:
	if err == nil {
		CreateQueue <- Request{RequestName: requestName, Workloads: deployments, RequestContext: context}
	} else {
		klog.Errorf("Error getting deployment and container labels from prometheus: %s", err)
		return
	}

	msg := fmt.Sprintf("Recommendation for cluster %s namespace %s   successfully submitted with recommendation Id %s", clusterName, nameSpace, requestName)
	_, err = w.Write([]byte(msg))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	klog.V(4).Info("Received recommendation request")

}
