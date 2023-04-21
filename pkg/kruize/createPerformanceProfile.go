package kruize

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/stolostron/recommends/pkg/utils"

	"github.com/stolostron/recommends/pkg/config"

	klog "k8s.io/klog/v2"
)

var create_performance_profile_url = config.Cfg.KruizeURL + "/createPerformanceProfile"

func InitPerformanceProfile() bool {

	postBody, err := os.ReadFile("./pkg/kruize/resource_optimization_ocm.json")
	if err != nil {
		klog.Errorf("Error reading resource_optimization_ocm.json: %v \n", err)
		return false
	}
	client := utils.HTTPClient()
	res, err := client.Post(create_performance_profile_url, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		klog.Errorf("Cannot create performanceprofile in kruize: %v \n", err)
		return false
	}
	if res.StatusCode == 201 {
		klog.Infof("Performance profile created successfully")
		return true
	}
	bodyBytes, _ := io.ReadAll(res.Body)
	data := map[string]interface{}{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		klog.Errorf("Cannot unmarshal response data: %v", err)
		return false
	}
	klog.V(2).Infof("Kruize response message : %v", data["message"])
	if data["message"] == "Validation failed: Performance Profile already exists: resource-optimization-acm" {
		klog.Infof("PerformanceProfile already exists")
		return true
	}
	klog.Infof("Kruize response code : %v", res.StatusCode)
	err = res.Body.Close()
	if err != nil {
		klog.Errorf("Cannot close response body: %v \n", err)
		return false
	}
	return false
}
