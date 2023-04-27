package server

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/kruize"
	"github.com/stolostron/recommends/pkg/utils"
	"k8s.io/klog"
)

var UpdateQueue chan CreateExperiment

func init() {
	UpdateQueue = make(chan CreateExperiment)
}

//gets values from getPerformanceProfileMetrics passed to it
//and passes them as request to updateResults kruize

var update_results_url = config.Cfg.KruizeURL + "/updateResults"

type UpdateResults struct {
	Version           string                    `json:"version"`
	ExperimentName    string                    `json:"experiment_name"`
	StartTimestamp    string                    `json:"start_timestamp"`
	EndTimestamp      string                    `json:"end_timestamp"`
	KubernetesObjects []KubernetesObjectMetrics `json:"kubernetes_objects"`
}

type KubernetesObjectMetrics struct {
	Type       string             `json:"type"`
	Name       string             `json:"name"`
	Namespace  string             `json:"namespace"`
	Containers []ContainerMetrics `json:"containers"`
}

type ContainerMetrics struct {
	ContainerImage string   `json:"container_image_name"`
	ContainerName  string   `json:"container_name"`
	Metrics        []Metric `json:"metrics"`
}

type Metric struct {
	Name    string `json:"name"`
	Results Result `json:"results"`
}

type Result struct {
	Value           float64               `json:"value,omitempty"`
	Format          string                `json:"format,omitempty"`
	AggregationInfo AggregationInfoValues `json:"aggregation_info"`
}

type AggregationInfoValues struct {
	AggregationInfo map[string]float64 `json:"aggregation_info"` //ex: "avg": 123.340
	Format          string             `json:"format"`
}

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

	var updateResultBody UpdateResults
	klog.V(5).Infof("Update Result Experiment: %s\n", ce.ExperimentName)
	pm := kruize.NewProfileManager("")
	for _, kubeobj := range ce.KubernetesObjects {
		for _, contlist := range kubeobj.Containers {

			// get queries from performanceProfile per container:
			metricsList := pm.GetPerformanceProfileInstanceMetrics(ce.ClusterName, kubeobj.Namespace,
				kubeobj.Name, contlist.ContainerName, ce.TrialSettings.MeasurementDuration)

			for _, metric := range metricsList {
				//call function to parse the metrics:
				updateResultBody = UpdateResults{
					Version:        ce.Version,
					ExperimentName: ce.ExperimentName,
					StartTimestamp: time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05"), //an hour ago from now
					EndTimestamp:   time.Now().Format("2006-01-02 15:04:05"),                      //now
					KubernetesObjects: []KubernetesObjectMetrics{
						{
							Type:      "deployment",
							Name:      kubeobj.Name,
							Namespace: kubeobj.Namespace,
							Containers: []ContainerMetrics{
								{
									ContainerImage: contlist.ContainerImage,
									ContainerName:  contlist.ContainerName,
									Metrics: []Metric{
										{
											Name: metric.Name,
											Results: Result{
												Value:  metric.Results.Value,
												Format: metric.Results.Format,
												AggregationInfo: AggregationInfoValues{
													AggregationInfo: metric.Results.AggregationInfo.AggregationInfo,
													Format:          metric.Results.Format,
												},
											},
										},
									},
								},
							},
						},
					},
				}

				requestBodies := []UpdateResults{updateResultBody}
				count := 0
				err := updateResult(requestBodies)
				for err != nil && count < config.Cfg.RetryCount {
					count = count + 1
					klog.Errorf("Cannot updateResult for createExperiment %s in kruize: Will retry \n", ce.ExperimentName)
					time.Sleep(time.Duration(config.Cfg.RetryInterval) * time.Millisecond)
					err = updateResult(requestBodies)
				}
				if err == nil {
					klog.Infof("UpdateResult for experiment %s created successfully", ce.ExperimentName)
				}
			}
		}

	}
}

// now := time.Now()
// val, _ := strconv.Atoi(strings.Split(ce.TrialSettings.MeasurementDuration, "m")[0])
// starttime := now.Add(-time.Duration(val) * time.Minute).Unix()

func updateResult(requestBodies []UpdateResults) error {
	requestBodyJSON, err := json.Marshal(requestBodies)
	if err != nil {
		klog.Error("Error encoding JSON:", err)
		return err
	}
	client := utils.HTTPClient()
	klog.V(5).Info("Posting updateResults to Kruize Service", bytes.NewBuffer(requestBodyJSON))
	res, err := client.Post(update_results_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	} else if res.StatusCode == 201 {
		return nil
	}
	return nil

}
