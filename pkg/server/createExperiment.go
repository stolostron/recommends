package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

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
	var containerDataClean []string
	var requestBody CreateExperiment
	var requestBodies []CreateExperiment

	containerMap := make(map[string][]Container)

	// parse the clusterID
	for name, id := range clusterID {
		reqBody.ExperimentName = fmt.Sprint(name + "-" + id)
		parts := strings.Split(name, "_")
		reqBody.ClusterName = parts[0]
		kubeObj.Namespace = parts[1]

	}

	//get containers
	for deployment, containerData := range deployments {
		containerDataClean = helpers.RemoveDuplicate(containerData)
		for _, contData := range containerDataClean {
			containerMap[deployment] = append(containerMap[deployment], Container{ContainerName: contData})
		}
	}

	for deployment, containers := range containerMap {
		for con := range containers {

			//parse deployment data
			requestBody = CreateExperiment{
				Version:            "v1",
				ExperimentName:     fmt.Sprintf("%s-%s-%d", reqBody.ExperimentName, deployment, con),
				ClusterName:        reqBody.ClusterName,
				PerformanceProfile: "resource_optimization_openshift",
				Mode:               "monitor",
				TargetCluster:      "local",
				KubernetesObjects: []KubernetesObject{
					{
						Type:       "deployment",
						Name:       deployment,
						Namespace:  kubeObj.Namespace,
						Containers: containers,
					},
				},
				TrialSettings: TrialSettings{
					MeasurementDuration: "15min",
				},
				RecommendationSettings: RecommendationSettings{
					Threshold: "0.1",
				},
			}

			requestBodies = append(requestBodies, requestBody)
			createExperiment(requestBodies, context)
		}

	}

}

func createExperiment(requestBodies []CreateExperiment, context context.Context) {

	//post createExperiment request to kruize
	requestBodyJSON, err := json.Marshal(requestBodies)
	if err != nil {
		klog.Error("Error encoding JSON:", err)
		return
	}
	client := utils.HTTPClient()
	// if post request fails retry with max wait time 30 sec
	retry := 0
	res, err := client.Post(create_experiment_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		// Max wait time is 30 sec
		timeToSleep := 30 * time.Second
		retry++
		klog.Errorf("Cannot create createExperiment %s in kruize: %v. Will retry in %s\n", requestBodies[0].ExperimentName, err, timeToSleep)
		time.Sleep(timeToSleep) //wait 30 seconds
	} else if res.StatusCode == 201 {
		klog.Infof("CreateExperiment profile created successfully")
		bodyBytes, _ := io.ReadAll(res.Body)
		data := map[string]interface{}{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			klog.Errorf("Cannot unmarshal response data: %v", err)
		}
	}

}
