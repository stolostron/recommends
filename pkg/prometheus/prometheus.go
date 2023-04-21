package prometheus

import (
	"context"
	"errors"
	"fmt"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stolostron/recommends/pkg/config"
	"k8s.io/klog"
)

type Result struct {
	Container    string `json:"container"`
	Workload     string `json:"workload"`
	WorkloadType string `json:"workloadType"`
}

func GetLabels(clusterName string, namespace string) (map[string][]string, error) {

	//setup context with a timeout to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	deploymentContainers := make(map[string][]string)

	client, err := promApi.NewClient(promApi.Config{
		Address: config.Cfg.ThanosURL,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %v. Please ensure that the API server is running and the address is correct", err)
	}

	v1api := promv1.NewAPI(client)
	clusterFilter := `cluster="` + clusterName + `"`
	namespaceFilter := `namespace="` + namespace + `"`
	allFilter := clusterFilter + `,` + namespaceFilter
	query := `sum(
		node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{` + allFilter + `}
	  * on(namespace, pod)
		group_left( workload_type, workload) namespace_workload_pod:kube_pod_owner:relabel{` + allFilter + `, workload_type=~".+"}
	) by (container,workload,workload_type)`

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
			workloadType := labels["workload_type"]

			r := Result{
				Container:    string(container),
				Workload:     string(workload),
				WorkloadType: string(workloadType),
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
