package utils

import (
	"crypto/tls"
	"net/http"

	promApi "github.com/prometheus/client_golang/api"
	"github.com/stolostron/recommends/pkg/config"
	"k8s.io/klog"
)

var PromClient promApi.Client

func HTTPClient() http.Client {
	// Create a transport configuration with TLS (HTTPS) settings
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	c := http.Client{
		Transport: tr,
	}
	return c
}

func init() {
	klog.Info("Initializing Prometheus client")
	client, err := promApi.NewClient(promApi.Config{
		Address: config.Cfg.ThanosURL,
	})
	if err != nil {
		klog.Errorf("Error creating client: %v. Please ensure that the Thanos server is running and the address is correct", err)
	}
	PromClient = client
}
