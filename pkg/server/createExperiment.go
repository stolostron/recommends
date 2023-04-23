package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	Version                string                 `json:"version"`
	ExperimentName         string                 `json:"experiment_name"`
	ClusterName            string                 `json:"cluster_name"`
	PerformanceProfile     string                 `json:"performance_profile"`
	Mode                   string                 `json:"mode"`
	TargetCluster          string                 `json:"target_cluster"`
	KubernetesObjects      []KubernetesObject     `json:"kubernetes_objects"`
	TrialSettings          TrialSettings          `json:"trial_settings"`
	RecommendationSettings RecommendationSettings `json:"recommendation_settings"`
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

func LoadValues(requestName string, deployments map[string][]string, context context.Context) {

	var reqBody CreateExperiment
	var kubeObj KubernetesObject
	var containerDataClean []string
	var requestBody CreateExperiment

	containerMap := make(map[string][]Container)

	//get containers
	for deployment, containerData := range deployments {
		containerDataClean = helpers.RemoveDuplicate(containerData)
		for _, contData := range containerDataClean {
			containerMap[deployment] = append(containerMap[deployment], Container{ContainerName: contData})
		}
	}

	for deployment, containers := range containerMap {
		for _, con := range containers {
			singleContainer := []Container{con}
			experimentName := fmt.Sprintf("%s-%s-%s", requestName, deployment, con.ContainerName)

			//parse deployment data
			requestBody = CreateExperiment{
				Version:            "1.0",
				ExperimentName:     experimentName,
				ClusterName:        reqBody.ClusterName,
				PerformanceProfile: "resource-optimization-acm",
				Mode:               "monitor",
				TargetCluster:      "remote",
				KubernetesObjects: []KubernetesObject{
					{
						Type:       "deployment",
						Name:       deployment,
						Namespace:  kubeObj.Namespace,
						Containers: singleContainer,
					},
				},
				TrialSettings: TrialSettings{
					MeasurementDuration: "60min",
				},
				RecommendationSettings: RecommendationSettings{
					Threshold: "0.1",
				},
			}

			requestBodies := []CreateExperiment{requestBody}
			count := 0
			err := createExperiment(requestBodies, context)
			for err != nil && count < config.Cfg.RetryCount {
				count = count + 1
				klog.Errorf("Cannot create createExperiment %s in kruize: Will retry \n", experimentName)
				time.Sleep(time.Duration(config.Cfg.RetryInterval) * time.Millisecond)
				err = createExperiment(requestBodies, context)
			}
			if err != nil {
				klog.Infof("CreateExperiment %s profile created successfully", experimentName)
			}

		}

	}

}

func createExperiment(requestBodies []CreateExperiment, context context.Context) error {
	//post createExperiment request to kruize
	requestBodyJSON, err := json.Marshal(requestBodies)
	if err != nil {
		klog.Error("Error encoding JSON:", err)
		return err
	}
	client := utils.HTTPClient()
	klog.Info("Posting create Experiment to Kruize Service", bytes.NewBuffer(requestBodyJSON))
	res, err := client.Post(create_experiment_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	} else if res.StatusCode == 201 {
		return nil
	}
	return nil

}
