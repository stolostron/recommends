package server

import (
	"time"

	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/kruize"
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
	ContainerImage string    `json:"container_image_name"`
	ContainerName  string    `json:"container_name"`
	Metrics        []Metrics `json:"metrics"`
}

type Metrics struct {
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
func updateResultRequest(ce *CreateExperiment) UpdateResults {

	var updateResult UpdateResults
	klog.V(5).Infof("Update Result Experiment: %s\n", ce.ExperimentName)
	pm := kruize.NewProfileManager("")
	for _, kubeobj := range ce.KubernetesObjects {
		for _, contlist := range kubeobj.Containers {

			// get queries from performanceProfile per container:
			metricsList := pm.GetPerformanceProfileInstanceMetrics(ce.ClusterName, kubeobj.Namespace,
				kubeobj.Name, contlist.ContainerName, ce.TrialSettings.MeasurementDuration)

			//call function to parse the metrics:
			starttime := time.Now().Unix()
			updateResult = &UpdateResults{
				Version:        ce.Version,
				ExperimentName: ce.ExperimentName,
				StartTimestamp: time.Unix(starttime, 0).Format("2006-01-02 15:04:05"), //an hour ago from now
				EndTimestamp:   time.Now().Format("2006-01-02 15:04:05"),              //now
				KubernetesObjects: []KubernetesObjectMetrics{
					{
						Type:      "deployment",
						Name:      kubeobj.Name,
						Namespace: kubeobj.Namespace,
						Containers: []ContainerMetrics{
							{
								ContainerImage: contlist.ContainerImage,
								ContainerName:  contlist.ContainerName,
								Metrics: []Metrics{
									metricsList,
								},
							},
						},
					},
				},
			},
			
		}

	}
	return updateResult
}


// now := time.Now()
// val, _ := strconv.Atoi(strings.Split(ce.TrialSettings.MeasurementDuration, "m")[0])
// starttime := now.Add(-time.Duration(val) * time.Minute).Unix()
