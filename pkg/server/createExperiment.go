package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/stolostron/recommends/pkg/helpers"
)

//reads in the values from computeRecommendations and passes them to request to createExperiment kruize

type RequestBody struct {
	Version                string             `json:"version"`
	ExperimentName         string             `json:"experiment_name"`
	ClusterName            string             `json:"cluster_name"`
	PerformanceProfile     string             `json:"performace_profile"`
	Mode                   string             `json:"mode"`
	TargetCluster          string             `json:"target_cluster"`
	KubernetesObjects      []KubernetesObject `json:"kubernetes_objects"` //this should be a list of deployments map[string][]string
	TrialSettings          TrialSettings
	RecommendationSettings RecommendationSettings
}

type KubernetesObject struct { //these are the values in the deployment
	Type       string      `json:"type"`
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Containers []Container `json:"containers"`
}

type Container struct { //these are the container values within the deployments (above)
	ContainerImage string `json:"container_image_name"`
	ContainerName  string `json:"container_name"`
}

type TrialSettings struct {
	MeasurementDuration string `json:"measurement_duration"`
}

type RecommendationSettings struct {
	Threshold string `json:"threshold"`
}

func LoadValues(clusterID map[string]string, deployments map[string][]string) {

	var responseBody RequestBody
	var kubeObj KubernetesObject
	var conObj Container
	var containerDataClean []string

	// parse the clusterID
	for name, id := range clusterID {
		responseBody.ExperimentName = fmt.Sprint(name + "-" + id)
		parts := strings.Split(name, "_")
		responseBody.ClusterName = parts[0]
		kubeObj.Namespace = parts[1]

	}
	//parse deployment data
	for deployment, containerData := range deployments {
		containerDataClean = helpers.RemoveDuplicate(containerData)
		fmt.Println("After Clean: ", deployment, containerDataClean)
		kubeObj.Name = deployment
		for _, contData := range containerDataClean {
			conObj = Container{ContainerImage: "", ContainerName: contData}
		}

	}

	//create createExperiement object TODO: make hard coded values env variables
	requestBody := RequestBody{
		Version:            "v1",
		ExperimentName:     responseBody.ExperimentName,
		ClusterName:        responseBody.ClusterName,
		PerformanceProfile: responseBody.PerformanceProfile,
		Mode:               responseBody.Mode,
		TargetCluster:      responseBody.TargetCluster,
		KubernetesObjects: []KubernetesObject{
			{
				Type:      "deployment",
				Name:      kubeObj.Name,
				Namespace: kubeObj.Namespace,
				Containers: []Container{
					{
						ContainerImage: conObj.ContainerImage,
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

	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", "https://localhost:8080/createExperiment", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// send the request and process the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()
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
