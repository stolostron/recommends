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
	Value           *float64              `json:"value,omitempty"`
	Format          string                `json:"format,omitempty"`
	AggregationInfo AggregationInfoStruct `json:"aggregation_info"`
}

type AggregationInfoStruct struct {
	Sum    *float64 `json:"sum"`
	Max    *float64 `json:"max,omitempty"`
	Min    *float64 `json:"min,omitempty"`
	Avg    *float64 `json:"avg"`
	Format string   `json:"format"`
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

	klog.V(5).Infof("Update Result Experiment: %s\n", ce.ExperimentName)

	// now := time.Now()
	// val, _ := strconv.Atoi(strings.Split(ce.TrialSettings.MeasurementDuration, "m")[0])
	// starttime := now.Add(-time.Duration(val) * time.Minute).Unix()

	pm := kruize.NewProfileManager("")

	// var containerMetrics ContainerMetrics
	for _, kubeobj := range ce.KubernetesObjects {
		for _, contlist := range kubeobj.Containers {

			// get queries from performanceProfile:
			queryNameMap := pm.GetPerformanceProfileInstance(ce.ClusterName, kubeobj.Namespace,
				kubeobj.Name, contlist.ContainerName, ce.TrialSettings.MeasurementDuration)

			// get metrics from perfprofile queries
			metrics := kruize.GetMetricsForQuery(queryNameMap)

			//call function to parse the metrics:
			starttime := time.Now().Unix()
			updateResult := &UpdateResults{
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
									{
										Name: metrics.Name,
										Results: Result{
											Value:  metrics.Results.Value,
											Format: metrics.Results.Format,
											AggregationInfo: AggregationInfoStruct{
												Avg:    metrics.Results.AggregationInfo.Avg,
												Max:    metrics.Results.AggregationInfo.Max,
												Min:    metrics.Results.AggregationInfo.Min,
												Sum:    metrics.Results.AggregationInfo.Sum,
												Format: metrics.Results.Format,
											},
										},
									},
								},
							},
						},
					},
				},
			}

			klog.V(5).Info(updateResult)
		}
	}

	// UpdateQueue <- updateResult

}

func getQueries() {

	// TODO: separate out the functions used in above func

}
