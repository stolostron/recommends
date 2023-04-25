package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/prometheus"
	"k8s.io/klog"
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

// prepares recommendation request
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
	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		var syntaxError *json.SyntaxError

		switch {

		case errors.As(err, &syntaxError):
			http.Error(w, "{\"message\":\"Request body contains badly-formed JSON.\"}", http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %s field", unmarshalTypeError.Field)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.EOF):
			http.Error(w, "{\"message\":\"Request body must not be empty.\"}", http.StatusBadRequest)

		default:
			klog.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
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
		uid := helpers.GenerateID(requestName)
		requestIdMap[requestName] = uid
		requestName = requestName + "_" + uid
	}

	//get the deployments and containers:
	deployments, err := prometheus.GetLabels(clusterName, nameSpace)

	//createExperiment with data:
	if err == nil {
		//LoadValues(requestName, deployments, context)
		CreateQueue <- Request{RequestName: requestName, Workloads: deployments, RequestContext: context}
	} else {
		klog.Errorf("Error getting deployment and container labels from prometheus: %s", err)
		return
	}
	//TODO: decide if we need this
	//append to recommendations list temporary store in memory
	//recommendations = append(recommendations, newRecommendation...)

	msg := fmt.Sprintf("Recommendation for cluster %s namespace %s   successfully submitted with recommendation Id %s", clusterName, nameSpace, requestName)
	_, err = w.Write([]byte(msg))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	klog.V(4).Info("Received recommendation request")

}
