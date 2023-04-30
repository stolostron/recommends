// Copyright Contributors to the Open Cluster Management project

package model

type Perf_profile struct {
	K8s_type        string `json:"k8s_type"`
	Name            string `json:"name"`
	Profile_version int    `json:"profile_version"`
	Slo             slo    `json:"slo"`
}

type slo struct {
	Direction          string              `json:"direction"`
	Function_variables []Function_variable `json:"function_variables"`
	Objective_function obj_function        `json:"objective_function"`
	Slo_class          string              `json:"slo_class"`
}

type Function_variable struct {
	Aggregation_functions []Aggregation_function `json:"aggregation_functions"`
	Datasource            string                 `json:"datasource"`
	Kubernetes_object     string                 `json:"kubernetes_object"`
	Name                  string                 `json:"name"`
	Value_type            string                 `json:"value_type"`
}

type Aggregation_function struct {
	Function string `json:"function"`
	Query    string `json:"query"`
	Versions string `json:"versions"`
}

type obj_function struct {
	Function_type string `json:"function_type"`
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
	AggregationInfo map[string]interface{} `json:"aggregation_info"`
}
