package server

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog"
)

type Result struct {
	// Pod       string `json:"pod"`
	Container string `json:"container"`
	Workload  string `json:"workload"`
}

var DepCon map[string][]string

func GetLabels() {
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

		r := Result{
			// Pod:          string(pod),
			Container: string(container),
			Workload:  string(workload),
		}

		if _, ok := deploymentContainers[r.Workload]; !ok {
			deploymentContainers[r.Workload] = make([]string, 0)
		}
		deploymentContainers[r.Workload] = append(deploymentContainers[r.Workload], r.Container)
	}

}
