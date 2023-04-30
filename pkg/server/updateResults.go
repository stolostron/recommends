package server

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/stolostron/recommends/pkg/kruize"
	"github.com/stolostron/recommends/pkg/model"
	klog "k8s.io/klog/v2"
)

var UpdateQueue chan CreateExperiment

type UpdateResults struct {
	Version           string                          `json:"version"`
	ExperimentName    string                          `json:"experiment_name"`
	StartTimestamp    string                          `json:"interval_start_time"`
	EndTimestamp      string                          `json:"interval_end_time"`
	KubernetesObjects []model.KubernetesObjectMetrics `json:"kubernetes_objects"`
}

func init() {
	UpdateQueue = make(chan CreateExperiment)
}

//var update_results_url = config.Cfg.KruizeURL + "/updateResults"

func ProcessUpdateQueue(q chan CreateExperiment) {
	for {
		klog.Info("Processing update Q")
		ce := <-q
		updateResultRequest(&ce)
		klog.Infof("Processed %s", ce.ExperimentName)
	}
}

//updateresults per each experiment
func updateResultRequest(ce *CreateExperiment) UpdateResults {

	var updateResult UpdateResults
	var updateResults []UpdateResults

	klog.V(5).Infof("Update Result Experiment: %s\n", ce.ExperimentName)
	pm := kruize.NewProfileManager("")
	for _, kubeobj := range ce.KubernetesObjects {
		for _, contlist := range kubeobj.Containers {

			// get queries from performanceProfile per container:
			metricsList := pm.GetPerformanceProfileInstanceMetrics(ce.ClusterName, kubeobj.Namespace,
				kubeobj.Name, contlist.ContainerName, ce.TrialSettings.MeasurementDuration)

			//call function to parse the metrics:
			windowSizeStr := strings.TrimSuffix(ce.TrialSettings.MeasurementDuration, "min")
			windowSizeInt, _ := strconv.ParseInt(windowSizeStr, 10, 64)
			startTime := time.Now()
			endTime := startTime.Add(time.Duration(windowSizeInt) * time.Minute)
			updateResult = UpdateResults{
				Version:        ce.Version,
				ExperimentName: ce.ExperimentName,
				StartTimestamp: startTime.Format("2006-01-02T15:04:05.000Z"), //an hour ago from now
				EndTimestamp:   endTime.Format("2006-01-02T15:04:05.000Z"),   //now
				KubernetesObjects: []model.KubernetesObjectMetrics{
					{
						Type:      "deployment",
						Name:      kubeobj.Name,
						Namespace: kubeobj.Namespace,
						Containers: []model.ContainerMetrics{
							{
								ContainerImage: contlist.ContainerImage,
								ContainerName:  contlist.ContainerName,
								Metrics:        metricsList,
							},
						},
					},
				},
			}
			updateResults = append(updateResults, updateResult)

		}

	}
	up, _ := json.Marshal(updateResults)
	klog.V(9).Infof("Created updateResults Object %v", string(up))
	return updateResult
}
