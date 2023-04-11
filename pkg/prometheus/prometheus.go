package prometheus

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog"
)

type Result struct {
	Container string `json:"container"`
	Workload  string `json:"workload"`
}

func GetLabels(ctx context.Context) (map[string][]string, error) {

	//set a timeout for the context to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	deploymentContainers := make(map[string][]string)

	client, err := api.NewClient(api.Config{
		Address: "http://localhost:5555",
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v. Please ensure that the API server is running and the address is correct", err)
	}

	v1api := v1.NewAPI(client)

	query := `sum(
		node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{cluster="local-cluster", namespace="open-cluster-management-observability"}
	  * on(namespace, pod)
		group_left( workload_type, workload) namespace_workload_pod:kube_pod_owner:relabel{cluster="local-cluster", namespace="open-cluster-management-observability", workload_type="deployment"}
	) by (container,workload)`

	res, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("API query timed out: %v", err)
		}
		return nil, fmt.Errorf("API query failed: %v", err)
	}

	vector := res.(model.Vector)
	for _, sample := range vector {
		if sample.Metric != nil {
			klog.V(4).Infof("Name: %s, Labels: %s", sample.Metric["__name__"], sample.Metric)
			labels := sample.Metric
			container := labels["container"]
			workload := labels["workload"]

			r := Result{
				Container: string(container),
				Workload:  string(workload),
			}

			if _, ok := deploymentContainers[r.Workload]; !ok {
				deploymentContainers[r.Workload] = make([]string, 0)
			}
			deploymentContainers[r.Workload] = append(deploymentContainers[r.Workload], r.Container)
		} else {
			return nil, fmt.Errorf("Metric results are empty. Please ensure query is correct and returns required labels: %s.", query)

		}
	}
	return deploymentContainers, nil
}
