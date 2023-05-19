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

type RecommendationStore struct {
	data []*RecommendationItem
}

var Recommendationstore RecommendationStore

func processRequest(req *Request) {
	klog.Infof("Processing Request %s", req.RequestName)
	var requestBody CreateExperiment
	var containerDataClean []string
	containerMap := make(map[string][]Container)
	containerObjectMap := make(map[string][]string)

	//get containers
	for deployment, containerData := range req.Workloads {
		containerDataClean = helpers.RemoveDuplicate(containerData)
		for _, contData := range containerDataClean {
			containerMap[deployment] = append(containerMap[deployment], Container{ContainerImage: contData, ContainerName: contData})
			containerObjectMap[deployment] = append(containerObjectMap[deployment], contData)
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

			clusterNamespaceMap := SaveRecommendationData(containerObjectMap, req)
			Recommendationstore.data = append(Recommendationstore.data, clusterNamespaceMap)

		}

		//Add break here to run one deployment for test
		// break
	}

	klog.V(5).Infof("Processed %s", req.RequestName)
}

func createExperiment(requestBodies []CreateExperiment, context context.Context) error {
	//post createExperiment request to kruize
	requestBodyJSON, err := json.Marshal(requestBodies)
	if err != nil {
		klog.Error("Error encoding JSON:", err)
		return err
	}
	client := utils.HTTPClient()
	klog.Infof("Creating experiment %s", requestBodies[0].ExperimentName)
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
		klog.V(9).Info("Processing create Q")
		req := <-q
		go processRequest(&req)
	}
}

func SaveRecommendationData(containerMapObject map[string][]string, req *Request) *RecommendationItem {

	var recommendationItem = RecommendationItem{}

	reqParts := strings.Split(req.RequestName, "_")

	recommendationItem.Cluster = reqParts[1]
	recommendationItem.Namespace = reqParts[2]
	recommendationItem.RecommendationID = reqParts[3]

	deployments := make(map[string][]map[string]string)

	for deployment, containers := range containerMapObject {
		containerRecommendations := make([]map[string]string, 0) //ex: [{cont1:rec1}, {con2:rec2},..]
		for _, container := range containers {
			containerRecommendation := map[string]string{
				container: "recommendation-status",
			}
			containerRecommendations = append(containerRecommendations, containerRecommendation)
		}

		deployments = map[string][]map[string]string{
			deployment: containerRecommendations,
		}
	}

	recommendationItem.Recommendation = deployments
	return &recommendationItem

}
