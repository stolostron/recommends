package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/utils"
	"k8s.io/klog"
)

//reads in the values from computeRecommendations and prometheus
//and passes them as request to createExperiment kruize

var create_experiment_url = config.Cfg.KruizeURL + "/createExperiment"

type CreateExperiment struct {
	Version                string             `json:"version"`
	ExperimentName         string             `json:"experiment_name"`
	ClusterName            string             `json:"cluster_name"`
	PerformanceProfile     string             `json:"performace_profile"`
	Mode                   string             `json:"mode"`
	TargetCluster          string             `json:"target_cluster"`
	KubernetesObjects      []KubernetesObject `json:"kubernetes_objects"`
	TrialSettings          TrialSettings
	RecommendationSettings RecommendationSettings
}

type KubernetesObject struct {
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Containers []Container `json:"containers"`
}

type Container struct {
	ContainerImage string `json:"container_image_name"`
	ContainerName  string `json:"container_name"`
}

type TrialSettings struct {
	MeasurementDuration string `json:"measurement_duration"`
}

type RecommendationSettings struct {
	Threshold string `json:"threshold"`
}

func LoadValues(clusterID map[string]string, deployments map[string][]string, context context.Context) {

	var reqBody CreateExperiment
	var kubeObj KubernetesObject
	var conObj Container
	var containerDataClean []string

	// parse the clusterID
	for name, id := range clusterID {
		reqBody.ExperimentName = fmt.Sprint(name + "-" + id)
		parts := strings.Split(name, "_")
		reqBody.ClusterName = parts[0]
		kubeObj.Namespace = parts[1]

	}
	//parse deployment data
	for deployment, containerData := range deployments {
		containerDataClean = helpers.RemoveDuplicate(containerData)
		kubeObj.Name = deployment
		for _, contData := range containerDataClean {
			conObj = Container{ContainerImage: "", ContainerName: contData}
		}

	}
	//create createExperiement object TODO: make hard coded values env variables
	requestBody := CreateExperiment{
		Version:            "v1",
		ExperimentName:     reqBody.ExperimentName,
		ClusterName:        reqBody.ClusterName,
		PerformanceProfile: "resource_optimization_openshift",
		Mode:               "monitor",
		TargetCluster:      "local",
		KubernetesObjects: []KubernetesObject{
			{
				Type:      "deployment",
				Name:      kubeObj.Name,
				Namespace: kubeObj.Namespace,
				Containers: []Container{
					{
						ContainerImage: "kruize/tfb-db:1.15",
						ContainerName:  conObj.ContainerName,
					},
				},
			},
		},
		TrialSettings: TrialSettings{
			MeasurementDuration: "15min",
		},
		RecommendationSettings: RecommendationSettings{
			Threshold: "0.1",
		},
	}
	requestBodies := []CreateExperiment{requestBody}

	//post createExperiment request to kruize
	requestBodyJSON, err := json.Marshal(requestBodies)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	klog.V(5).Info("%+v\n", requestBody)
	client := utils.HTTPClient()
	res, err := client.Post(create_experiment_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		klog.Errorf("Cannot create createExperiment with context %s in kruize: %v \n", context, err)
		klog.Info(res)
		return
	}
	if res.StatusCode == 201 {
		klog.Infof("CreateExperiment profile created successfully")
		return
	}
	bodyBytes, _ := io.ReadAll(res.Body)
	data := map[string]interface{}{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		klog.Errorf("Cannot unmarshal response data: %v", err)
		return
	}
}

// sample of createExperiment:
// [{
// 	"version": "1.0",
// 	"experiment_name": "quarkus-resteasy-autotune-min-http-response-time-db-new-1",
// 	"cluster_name": "cluster-one-division-bell",
// 	"performance_profile": "resource-optimization-openshift",
// 	"mode": "monitor",
// 	"target_cluster": "local",
// 	"kubernetes_objects": [
// 	  {
// 		"type": "deployment",
// 		"name": "tfb-qrh-deployment",
// 		"namespace": "default",
// 		"containers": [
// 		  {
// 			"container_image_name": "kruize/tfb-db:1.15",
// 			"container_name": "tfb-server-0"
// 		  },
// 		  {
// 			"container_image_name": "kruize/tfb-qrh:1.13.2.F_et17",
// 			"container_name": "tfb-server-1"
// 		  }
// 		]
// 	  }
// 	],
// 	"trial_settings": {
// 	  "measurement_duration": "15min"
// 	},
// 	"recommendation_settings": {
// 	  "threshold": "0.1"
// 	}
//   }]
