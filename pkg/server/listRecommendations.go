package server

import (
	"encoding/json"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/helpers"
	"github.com/stolostron/recommends/pkg/utils"

	klog "k8s.io/klog/v2"
)

var list_recommendations_url = config.Cfg.KruizeURL + "/listRecommendations"

type GetRecommendations struct {
	RecommendationId string `json:"recommendationId"`
	Namespace        string `json:"namespace"`
}

type KruizeRecommendations struct {
	recommendations map[string][]ContainerRecommendations

	// recommendenationID: [
	//  container1: status
	//	contaienr2: status
	// ]

}

type ContainerRecommendations struct {
	ContainerName  string
	ExperimentName string
	Status         string
}

var GetRecommendationsQueue chan GetRecommendations

func init() {
	GetRecommendationsQueue = make(chan GetRecommendations)
}

func listRecommendations(w http.ResponseWriter, r *http.Request) {
	klog.Infof("Received Request for list Recommendations")

	var getRecommendations GetRecommendations
	//create context from request

	// context := r.Context()

	//check content type is json
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&getRecommendations)

	if ok := helpers.ErrorHandlingRequests(w, err); !ok {
		return
	}

	recommednationId := getRecommendations.RecommendationId
	namespace := getRecommendations.Namespace
	//from the request get the experiment name and send request to kruize to get the recommendatrions

	var use string

	if namespace != "" && recommednationId == "" {
		use = namespace
	} else if namespace == "" && recommednationId != "" {
		use = recommednationId
		// use = strings.Split(use, "_")[2]
		klog.Info(use)
	}

	//for each namespace - we want to send listRecommendation request to Kruize
	// and save the result in an object

	//call kruize listRecommendation

	//ex: http://<ip>:<kruize port>/listRecommendations?experiment_name=ns_local-cluster_open-cluster-management-observability_e2389b95b9e70ccda0b44d096f10fb29ae125de9a97b62d959d841409dea1c7b-observability-alertmanager-config-reloader
	client := utils.HTTPClient()

	klog.Info("Request: ", list_recommendations_url+"?"+recommednationId)

	resp, err := client.Get(list_recommendations_url + recommednationId)

	if err != nil {
		klog.Error(err)

	} else if resp.StatusCode == 201 {
		klog.Info("Successfully got recommendations!")
	}

}
