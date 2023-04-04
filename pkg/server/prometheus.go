package server

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Result struct { //might not need:
	Pod          string  `json:"pod"`
	Container    string  `json:"container"`
	WorkloadType string  `json:"workload_type"`
	Workload     string  `json:"workload"`
	Value        float64 `json:"value"`
}

var DepCon map[string][]string

func PrometheusClient() {
	// Create a new Prometheus API client.
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:5555",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	v1api := v1.NewAPI(client)

	query := `sum(
		node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster="local-cluster", namespace="open-cluster-management-observability"}
	  * on(namespace,pod)
		group_left( workload_type, workload) namespace_workload_pod:kube_pod_owner:relabel{cluster="local-cluster", namespace="open-cluster-management-observability", workload_type="deployment"}
	) by (pod,container,workload_type, workload)`

	res, _, err := v1api.Query(context.Background(), query, time.Now())
	if err != nil {
		panic(err)
	}
	var results []Result

	vector := res.(model.Vector)
	for _, sample := range vector {
		fmt.Printf("Name: %s, Labels: %v,", sample.Metric["__name__"], sample.Metric)
		labels := sample.Metric
		container := labels["container"]
		pod := labels["pod"]
		workloadType := labels["workload_type"]
		workload := labels["workload"]

		r := Result{
			Pod:          string(pod),
			Container:    string(container),
			WorkloadType: string(workloadType),
			Workload:     string(workload),
		}

		results = append(results, r)
	}

	fmt.Println(results)

}
