package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/utils"
	klog "k8s.io/klog/v2"
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

func processRequest(req *Request) {
	klog.Infof("Processing Request %s", len(req.RequestName))
	var requestBody CreateExperiment
	var containerDataClean []string

	containerMap := make(map[string][]Container)

	//get containers
	for deployment, containerData := range req.Workloads {
		containerDataClean = helpers.RemoveDuplicate(containerData)
		for _, contData := range containerDataClean {
			containerMap[deployment] = append(containerMap[deployment], Container{ContainerImage: contData, ContainerName: contData})
		}
	}

	for deployment, containers := range containerMap {
		for _, con := range containers {
			singleContainer := []Container{con}
			experimentName := fmt.Sprintf("%s-%s-%s", req.RequestName, deployment, con.ContainerName)
			clusterName := strings.Split(req.RequestName, "_")[1]
			namespace := strings.Split(req.RequestName, "_")[2]
			requestBody = CreateExperiment{
				Version:            "1.0",
				ExperimentName:     experimentName,
				ClusterName:        clusterName,
				PerformanceProfile: "resource-optimization-acm",
				Mode:               "monitor",
				TargetCluster:      "remote",
				KubernetesObjects: []KubernetesObject{
					{
						Type:       "deployment",
						Name:       deployment,
						Namespace:  namespace,
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
			err := createExperiment(requestBodies, req.RequestContext)
			for err != nil && count < config.Cfg.RetryCount {
				count = count + 1
				klog.Errorf("Cannot create createExperiment %s in kruize: Will retry \n", experimentName)
				time.Sleep(time.Duration(config.Cfg.RetryInterval) * time.Millisecond)
				err = createExperiment(requestBodies, req.RequestContext)
			}
			if err == nil {
				klog.V(5).Infof("CreateExperiment %s profile created successfully", experimentName)
				UpdateQueue <- requestBody
			}

		}
		// TODO Remove this block to iterate all Containers .
		break
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
	klog.V(5).Info("Posting create Experiment to Kruize Service", bytes.NewBuffer(requestBodyJSON))
	res, err := client.Post(create_experiment_url, "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return err
	} else if res.StatusCode == 201 {
		return nil
	}
	return nil

}
func ProcessCreateQueue(q chan Request) {
	for {
		klog.V(5).Info("Processing create Q")
		req := <-q
		processRequest(&req)
		klog.V(5).Infof("Processed %s", req.RequestName)
	}
}
