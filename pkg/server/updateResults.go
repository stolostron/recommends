package server

import (
	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/kruize"
	"k8s.io/klog"
)

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
	Value           string          `json:"value"`
	Format          string          `json:"format"`
	AggregationInfo AggregationInfo `json:"aggregation_info"`
}

type AggregationInfo struct {
	Sum    string `json:"sum"`
	Avg    string `json:"avg"`
	Format string `json:"format"`
}

func UpdateResultRequest(ce CreateExperiment, metrics kruize.Metrics) {

	klog.Infof("Experiment: %s\n", ce)
	klog.Infof("With metrics: %s\n", metrics)

}
