package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/golang/gddo/httputil/header"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stolostron/recommends/pkg/helpers"
	"k8s.io/klog"
)

type recommendation []struct {
	ClusterName         string `json:"clusterName"`
	Namespace           string `json:"namespace"`
	Application         string `json:"application"`
	MeasurementDuration string `json:"measurement_duration"` //ex: "15min"
}

type result struct {
	// Pod       string `json:"pod"`
	Container string `json:"container"`
	Workload  string `json:"workload"`
}

// adds an recommendation from JSON received in the request body.
func computeRecommendations(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Inside Compute Recommendations")

	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var newRecommendation recommendation

	err := dec.Decode(&newRecommendation)

	//error handling:
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
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	//get the clusterID (cluster name with namespace or applicaiton):
	clusterID := make(map[string]string) //ex: clustername-namespace:"id-12345"
	var concat string

	//use namespace:
	if newRecommendation[0].Application == "" && newRecommendation[0].Namespace != "" {
		concat = fmt.Sprintf("%s_%s", newRecommendation[0].ClusterName, newRecommendation[0].Namespace)
	}
	//use application
	if newRecommendation[0].Application != "" && newRecommendation[0].Namespace == "" {
		concat = fmt.Sprintf("%s_%s", newRecommendation[0].ClusterName, newRecommendation[0].Application)

		// if both applications and namespace is empty return
	} else if newRecommendation[0].Application == "" && newRecommendation[0].Namespace == "" {
		klog.V(4).Info("Request missing both Application and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both Application and Namespace cannot be empty.\"}", http.StatusBadRequest)
		return
	}

	//if id for clusterName already exists then we don't need to generate new oneß
	if clusterID[concat] == "" {
		clusterID[concat] = helpers.GenerateID(clusterID)

	}
	//TODO: decide if we need this
	//append to recommendations list temporary store in memory
	//recommendations = append(recommendations, newRecommendation...)

	msg := fmt.Sprintf("Recommendation for clusterID %s successfully submitted.", clusterID)
	_, err = w.Write([]byte(msg))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//get the deployments and containers:
	deployments := GetLabels()

	//createExperiment with data:
	LoadValues(clusterID, deployments)

	klog.V(4).Info("Received recommendation request")

}

func GetLabels() map[string][]string {
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:5555",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return nil
	}

	v1api := v1.NewAPI(client)

	query := `sum(
		node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster="local-cluster", namespace="open-cluster-management-observability"}
	  * on(namespace,pod)
		group_left( workload_type, workload) namespace_workload_pod:kube_pod_owner:relabel{cluster="local-cluster", namespace="open-cluster-management-observability", workload_type="deployment"}
	) by (pod, container, workload)` //do we need pod ?

	res, _, err := v1api.Query(context.Background(), query, time.Now())
	if err != nil {
		panic(err)
	}
	deploymentContainers := make(map[string][]string)

	vector := res.(model.Vector)
	for _, sample := range vector {
		klog.V(5).Info("Name: %s, Labels: %v,\n", sample.Metric["__name__"], sample.Metric)
		labels := sample.Metric
		// pod := labels["pod"]
		container := labels["container"]
		workload := labels["workload"]

		r := result{
			// Pod:          string(pod),
			Container: string(container),
			Workload:  string(workload),
		}

		if _, ok := deploymentContainers[r.Workload]; !ok {
			deploymentContainers[r.Workload] = make([]string, 0)
		}
		deploymentContainers[r.Workload] = append(deploymentContainers[r.Workload], r.Container)
	}
	return deploymentContainers
}
