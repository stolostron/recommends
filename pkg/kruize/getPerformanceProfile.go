package kruize

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

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
func (p *profileManager) GetPerformanceProfileInstance(clusterName string, namespace string,
	workloadName string, containerName string, measurementDur string) map[string][]string {
	instanceProfile := *p.performanceProfile

	queryNameMap := make(map[string][]string) //query: name
	var queryList []string

	for i, fv := range instanceProfile.Slo.Function_variables {
		for j, af := range fv.Aggregation_functions {

			af.Query = replaceTemplate(fv.Name, af.Function, af.Query, clusterName, namespace, workloadName, containerName, measurementDur)
			klog.V(9).Info("Updated aggregate function ", j)
			queryList = append(queryList, af.Query)
			queryNameMap[fv.Name] = append(queryList, queryList...)

		}
		klog.V(9).Info("Updated function variable ", i)

	}
	return queryNameMap

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
