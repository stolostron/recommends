package server

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/kruize"
	"github.com/stolostron/recommends/pkg/model"
	"github.com/stolostron/recommends/pkg/utils"
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

var update_results_url = config.Cfg.KruizeURL + "/updateResults"

func ProcessUpdateQueue(q chan CreateExperiment) {
	for {
		klog.Info("Processing update Q")
		ce := <-q
		updateResultRequest(&ce)
		klog.Infof("Processed %s", ce.ExperimentName)
	}
}

//updateresults per each experiment
func updateResultRequest(ce *CreateExperiment) {

	var updateResult UpdateResults
	var updateResults []UpdateResults

	klog.V(5).Infof("Update Result Experiment: %s\n", ce.ExperimentName)
	pm := kruize.NewProfileManager("")
	for _, kubeobj := range ce.KubernetesObjects {
		for _, contlist := range kubeobj.Containers {
			//call function to parse the metrics:
			windowSizeStr := strings.TrimSuffix(ce.TrialSettings.MeasurementDuration, "min")
			windowSizeInt, _ := strconv.ParseInt(windowSizeStr, 10, 64)

			endTimes := getTimeWindows(1, int(windowSizeInt))
			for i := 1; i < len(endTimes); i++ {
				startTime := endTimes[i-1]
				endTime := endTimes[i]
				unixTime := endTime.Unix()
				klog.Info("Unixtime %v", unixTime)
				// get queries from performanceProfile per container:
				metricsList := pm.GetPerformanceProfileInstanceMetrics(ce.ClusterName, kubeobj.Namespace,
					kubeobj.Name, contlist.ContainerName, ce.TrialSettings.MeasurementDuration, unixTime)

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
				updateResults = []UpdateResults{updateResult}
				up, _ := json.Marshal(updateResults)
				klog.V(9).Infof("Created updateResults Object %v", string(up))
				postUpdateResult(updateResults)
				time.Sleep(1 * time.Millisecond)
			}

		}

	}

}

func postUpdateResult(updateResults []UpdateResults) error {
	//post updateResults request to kruize
	requestBodyJSON, err := json.Marshal(updateResults)
	if err != nil {
		klog.Error("Error encoding JSON:", err)
		return err
	}
	client := utils.HTTPClient()
	klog.V(5).Info("Posting updateResult to Kruize Service", bytes.NewBuffer(requestBodyJSON))
	res, err := client.Post(update_results_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	} else if res.StatusCode == 201 {
		klog.V(5).Info("Successful updateResults reqest for request %s , startTime: %s , endTime: %s", updateResults[0].ExperimentName, updateResults[0].StartTimestamp, updateResults[0].ExperimentName, updateResults[0].EndTimestamp)
		return nil
	}
	return nil

}

func getTimeWindows(days int, windowSize int) []time.Time {
	var windows []time.Time
	currentTime := time.Now()                  // To Do : we should get the time from the input
	startTime := currentTime.AddDate(0, 0, -2) // Goback to one day
	startTime = startTime.Add(-time.Hour * 1)  // Kruize needs 24 + 1 data points
	windows = append(windows, startTime)

	for currentTime.After(startTime) {
		startTime = startTime.Add(time.Duration(windowSize) * time.Minute)
		windows = append(windows, startTime)
	}
	if len(windows)%2 != 0 {
		klog.Errorf("Error computing Windows , unexpected length %v", windows)
		windows = windows[:len(windows)-1]
	}
	return windows
}
