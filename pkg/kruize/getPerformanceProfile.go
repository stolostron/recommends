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

//gets perfprof per container and returns query and name map
func (p *profileManager) GetPerformanceProfileInstanceMetrics(clusterName string, namespace string,
	workloadName string, containerName string, measurementDur string) []Metrics {
	instanceProfile := *p.performanceProfile
	var metric Metrics
	var metricsList []Metrics
	// var aggregationInfoList []map[string]float64

	for _, fv := range instanceProfile.Slo.Function_variables {

		var format, function string
		var value float64
		aggregateInfo := make(map[string]float64)
		metric.Name = fv.Name

		if strings.Contains(fv.Name, "cpu") {
			format = "cores"
		} else {
			format = "MiB"
		}

		for _, af := range fv.Aggregation_functions {

			af.Query = replaceTemplate(fv.Name, af.Function, af.Query, clusterName, namespace, workloadName, containerName, measurementDur)
			value = getResults(af.Query)

			if format == "cores" {
				value = helpers.ConvertCpuUsageToCores(value)
			} else {
				value = helpers.ConvertMemoryUsageToMiB(value)
			}

			function = af.Function

			aggregateInfo[function] = value

		}
		klog.Info(aggregateInfo)

		metric.Results = Result{Value: value, Format: format, AggregationInfo: AggregationInfoValues{
			AggregationInfo: aggregateInfo,
			Format:          format,
		},
		}

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
	klog.Info(len(result.Slo.Function_variables))
	return result, true
}
