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
		klog.V(9).Info("Processing update Q")
		ce := <-q
		go updateResultRequest(&ce)
	}
}

//updateresults per each experiment
func updateResultRequest(ce *CreateExperiment) {

	metricsSent := make(map[string][]model.Metrics)
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
				// get queries from performanceProfile per container:
				metricsList := pm.GetPerformanceProfileInstanceMetrics(ce.ClusterName, kubeobj.Namespace,
					kubeobj.Name, contlist.ContainerName, ce.TrialSettings.MeasurementDuration, unixTime)

				metricsSent[contlist.ContainerName] = metricsList

				updateResult = UpdateResults{
					Version:        ce.Version,
					ExperimentName: ce.ExperimentName,
					StartTimestamp: startTime.Format("2006-01-02T15:04:05.000Z"), // current time -1 hour
					EndTimestamp:   endTime.Format("2006-01-02T15:04:05.000Z"),   // current Time
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
				upJson, _ := json.Marshal(updateResults)
				klog.V(9).Infof("Created updateResults Object %s", string(upJson))
				err := postUpdateResult(updateResults)
				count := 0
				for err != nil && count < config.Cfg.RetryCount {
					count = count + 1
					retryWait := time.Duration(config.Cfg.RetryInterval) * time.Millisecond
					klog.Errorf("Cannot post updateResult %s in kruize: Will retry in %d ms. \n", ce.ExperimentName, retryWait)
					time.Sleep(retryWait)
					err = postUpdateResult(updateResults)
					if err != nil {
						klog.Errorf("Error on updateResult %s in kruize:n", err.Error())
					}
				}

			}

		}

	}

	klog.Info("Container metrics map:", metricsSent)
}
func postUpdateResult(updateResults []UpdateResults) error {
	//post updateResults request to kruize
	requestBodyJSON, err := json.Marshal(updateResults)
	if err != nil {
		klog.Error("Error encoding JSON:", err)
		return err
	}
	client := utils.HTTPClient()
	klog.V(5).Infof("Posting updateResult %s", updateResults[0].ExperimentName)
	klog.V(9).Infof("Posting updateResult to Kruize Service %v", bytes.NewBuffer(requestBodyJSON))
	res, err := client.Post(update_results_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	} else if res.StatusCode == 201 {
		klog.V(5).Infof("Successful updateResults request for request %s , startTime: %s , endTime:%s", updateResults[0].ExperimentName, updateResults[0].StartTimestamp, updateResults[0].EndTimestamp)
	} else {
		klog.Warningf("Received unexpected status code(%s) from update request.", res.StatusCode)
	}
	return nil

}

func getTimeWindows(days int, windowSize int) []time.Time {
	var windows []time.Time
	currentTime := time.Now()                  // To Do : we should get the time from the input
	startTime := currentTime.AddDate(0, 0, -5) // Goback to one day
	startTime = startTime.Add(-time.Hour * 2)  // Kruize needs 24 + 2 data points
	windows = append(windows, startTime)

	for currentTime.After(startTime) {
		startTime = startTime.Add(time.Duration(windowSize) * time.Minute)
		windows = append(windows, startTime)
	}
	return windows
}
