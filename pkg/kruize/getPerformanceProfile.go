package kruize

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/model"
	klog "k8s.io/klog/v2"
)

type profileManager struct {
	performanceProfile *model.Perf_profile
}

func NewProfileManager(profile_name string) *profileManager {
	pm := &profileManager{}
	if pf, good := getPerformanceProfile(profile_name); good {
		pm.performanceProfile = &pf
	}
	return pm
}

//gets perfprof per container from Thanos and returns array of metrics
func (p *profileManager) GetPerformanceProfileInstanceMetrics(clusterName string, namespace string,
	workloadName string, containerName string, measurementDur string) []model.Metrics {
	instanceProfile := *p.performanceProfile
	var metric model.Metrics
	var metricsList []model.Metrics
	measurementDur = strings.TrimSuffix(measurementDur, "in")
	/* Iterate the following json object to form the Metrics object
		        "function_variables": [
	            {
	                "name": "cpuRequest",
	                "datasource": "prometheus",
	                "value_type": "double",
	                "kubernetes_object": "container",
	                "aggregation_functions": [
	                    {
	                        "function": "avg",
	                        "query": "avg(kube_pod_container_resource_requests{cluster=\"$CLUSTER_NAME$\",namespace=\"$NAMESPACE$\",pod=~\"$WORKLOAD_NAME$-[^-]*-[^-]*\",container=\"$CONTAINER_NAME$\"})"
	                    },
	                    {
	                        "function": "sum",
	                        "query": "sum(kube_pod_container_resource_requests{cluster=\"$CLUSTER_NAME$\",namespace=\"$NAMESPACE$\",pod=~\"$WORKLOAD_NAME$-[^-]*-[^-]*\",container=\"$CONTAINER_NAME$\"})"
	                    }
	                ]
	            },
	*/
	for _, fv := range instanceProfile.Slo.Function_variables {
		var format, function string
		aggregateInfo := make(map[string]interface{})
		metric.Name = fv.Name

		if strings.Contains(fv.Name, "cpu") {
			format = "cores"
		} else {
			format = "MiB"
		}
		aggregateInfo["format"] = format
		for _, af := range fv.Aggregation_functions {
			af.Query = replaceTemplate(fv.Name, af.Function, af.Query, clusterName, namespace, workloadName, containerName, measurementDur)
			value, err := getResults(af.Query)
			if err != nil {
				klog.V(5).Infof("Error running query %s", af.Query)
				continue
			}
			if format == "cores" {
				value = helpers.ConvertCpuUsageToCores(value)
			} else {
				value = helpers.ConvertMemoryUsageToMiB(value)
			}
			function = af.Function
			aggregateInfo[function] = value
		}
		klog.V(9).Info(aggregateInfo)
		metric.Results = model.Result{AggregationInfo: aggregateInfo}
		metricsList = append(metricsList, metric)
	}
	return metricsList
}

func replaceTemplate(name string, function string, query string, clusterName string, namespace string,
	workloadName string, containerName string, measurementDur string) string {

	klog.V(8).Infof("Template Query " + query)
	query = strings.ReplaceAll(query, "$CLUSTER_NAME$", clusterName)
	query = strings.ReplaceAll(query, "$NAMESPACE$", namespace)
	query = strings.ReplaceAll(query, "$WORKLOAD_NAME$", workloadName)
	query = strings.ReplaceAll(query, "$CONTAINER_NAME$", containerName)
	query = strings.ReplaceAll(query, "$MEASUREMENT_DURATION$", measurementDur)
	klog.V(8).Infof("Instance Query " + query)

	return query
}

func getPerformanceProfile(profileName string) (model.Perf_profile, bool) {
	var result model.Perf_profile
	defaultProfile := "./pkg/kruize/resource_optimization_ocm.json"
	if profileName == "" {
		profileName = defaultProfile
	} else {
		profileName = "./pkg/kruize/" + profileName + ".json"
	}
	json_file, err := os.Open(filepath.Clean(profileName))

	if err != nil {
		klog.Errorf("Error reading file %s : %v \n", profileName, err)
	}
	byteArray, err := io.ReadAll(json_file)

	if err != nil {
		klog.Errorf("Error reading performance profile %s : %v \n", profileName, err)
		return result, false
	}
	err = json.Unmarshal(byteArray, &result)

	if err != nil {
		klog.Errorf("Error reading performance profile %s : %v \n", profileName, err)
		return result, false
	}
	klog.Infof("SLO.Function_variables: %d", len(result.Slo.Function_variables))
	return result, true
}
