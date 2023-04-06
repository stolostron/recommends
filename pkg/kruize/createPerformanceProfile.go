package kruize

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/stolostron/recommends/pkg/config"

	klog "k8s.io/klog/v2"
)

func InitPerformanceProfile() bool {
	create_performance_profile_url := config.Cfg.KruizeURL + "/createPerformanceProfile"
	postBody, err := os.ReadFile("./pkg/kruize/resource_optimization_openshift.json")
	if err != nil {
		klog.Errorf("Error reading resource_optimization_openshift.json: %v \n", err)
		return false
	}
	res, err := http.Post(create_performance_profile_url, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		klog.Errorf("Cannot create performanceprofile in kruize: %v \n", err)
		return false
	}
	if res.StatusCode == 201 {
		klog.Infof("Performance profile created successfully")
		return true
	}
	defer res.Body.Close()
	bodyBytes, _ := io.ReadAll(res.Body)
	data := map[string]interface{}{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		klog.Errorf("Cannot unmarshal response data: %v", err)
		return false
	}
	klog.V(2).Infof("Kruize response message : %v", data["message"])
	if data["message"] == "Validation failed: Performance Profile already exists: resource-optimization-openshift" {
		klog.Infof("PerformanceProfile already exists")
		return true
	}
	klog.Infof("Kruize response code : %v", res.StatusCode)
	return false
}
