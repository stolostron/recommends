package server

import (
	"time"
)

// struct to capture response from kruize
type ListRecommendations struct {
	Cluster_name       string             `json:"cluster_name,omitempty"`
	Experiment_name    string             `json:"experiment_name,omitempty"`
	Version            string             `json:"version,omitempty"`
	Kubernetes_objects []kubernetesObject `json:"kubernetes_objects,omitempty"`
}

type kubernetesObject struct {
	K8stype    string      `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Namespace  string      `json:"namespace,omitempty"`
	Containers []container `json:"containers,omitempty"`
}

type container struct {
	Container_image_name string            `json:"container_image_name,omitempty"`
	Container_name       string            `json:"container_name,omitempty"`
	Metrics              []metric          `json:"metrics,omitempty"`
	Recommendations      NewRecommendation `json:"recommendations,omitempty"`
}

type metric struct {
	Name    string `json:"name,omitempty"`
	Results result `json:"results,omitempty"`
}

type result struct {
	Aggregation_info aggregation_info `json:"aggregation_info,omitempty"`
}

type aggregation_info struct {
	Min    string `json:"min,omitempty"`
	Max    string `json:"max,omitempty"`
	Sum    string `json:"sum,omitempty"`
	Avg    string `json:"avg,omitempty"`
	Format string `json:"format,omitempty"`
}

type NewRecommendation struct {
	Data          map[string]recommendationType `json:"data,omitempty"`
	Notifications map[string]notification       `json:"notifications,omitempty"`
}

type notification struct {
	NotifyType string `json:"type,omitempty"`
	Message    string `json:"message,omitempty"`
}

type recommendationType struct {
	Duration_based termbased `json:"duration_based,omitempty"`
}

type termbased struct {
	Short_term  recommendationObject `json:"short_term,omitempty"`
	Medium_term recommendationObject `json:"medium_term,omitempty"`
	Long_term   recommendationObject `json:"long_term,omitempty"`
}

type recommendationObject struct {
	Monitoring_start_time time.Time       `json:"monitoring_start_time,omitempty"`
	Monitoring_end_time   time.Time       `json:"monitoring_end_time,omitempty"`
	Duration_in_hours     float64         `json:"duration_in_hours,omitempty"`
	Pods_count            int             `json:"pods_count,omitempty"`
	Confidence_level      float64         `json:"confidence_level,omitempty"`
	Config                ConfigObject    `json:"config,omitempty"`
	Variation             ConfigObject    `json:"variation,omitempty"`
	Notifications         []notifications `json:"notifications,omitempty"`
}

type ConfigObject struct {
	Limits   recommendedConfig `json:"limits,omitempty"`
	Requests recommendedConfig `json:"requests,omitempty"`
}

type recommendedConfig struct {
	Cpu    recommendedValues `json:"cpu,omitempty"`
	Memory recommendedValues `json:"memory,omitempty"`
}

type recommendedValues struct {
	Amount float64 `json:"amount,omitempty"`
	Format string  `json:"format,omitempty"`
}

type notifications struct {
	Notificationtype string `json:"type,omitempty"`
	Message          string `json:"message,omitempty"`
}
