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

// create instance of NamespaceClusterID
var NcID = NamespaceClusterID{
	NamespaceClusters: make(map[string]NamespaceCluster),
}

func processRequest(req *Request) {
	klog.Infof("Processing Request %s", req.RequestName)
	var requestBody CreateExperiment
	var containerDataClean []string
	var status string
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
				status = fmt.Sprintf("Error Cannot create createExperiment %s", experimentName)
				count = count + 1
				klog.Errorf("Cannot create createExperiment %s in kruize: Will retry \n", experimentName)
				time.Sleep(time.Duration(config.Cfg.RetryInterval) * time.Millisecond)
				err = createExperiment(requestBodies, req.RequestContext)
			}
			if err == nil {
				status = "Good"
				klog.V(5).Infof("CreateExperiment %s profile created successfully", experimentName)
				UpdateQueue <- requestBody
			}

			NcID := SaveRecommendationData(deployment, con, req, status) //TODO:ADD THE DEPLOYMENT

			klog.Info(NcID)
		}
		//Add break here to run one deployment for test
		break
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

//save the recomendationid and cooresponding containers
func SaveRecommendationData(deployment string, con Container, req *Request, status string) *NamespaceClusterID {

	containerStatus := make(map[string]string)
	containerStatus[con.ContainerName] = status

	dep := Deployment{
		ContainerStatuses: []map[string]string{
			containerStatus,
		},
	}

	rec := Recommendation{
		Deployments: map[string]Deployment{
			deployment: dep,
		},
	}

	// Create a new NamespaceCluster object
	nc := NamespaceCluster{
		Recommendations: map[string]Recommendation{
			strings.Split(req.RequestName, "_")[3]: rec,
		},
	}

	reqParts := strings.Split(req.RequestName, "_")
	clusterNamespace := fmt.Sprintf("%s_%s", reqParts[1], reqParts[2])
	NcID.NamespaceClusters[clusterNamespace] = nc

	return &NcID

}
