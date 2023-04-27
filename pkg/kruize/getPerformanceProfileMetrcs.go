package kruize

import (
	"context"
	"errors"
	"time"

	promApi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stolostron/recommends/pkg/config"

	klog "k8s.io/klog/v2"
)

type Metrics struct {
	Name    string `json:"name"`
	Results Result `json:"results"`
}

type Result struct {
	Value           float64               `json:"value,omitempty"`
	Format          string                `json:"format"`
	AggregationInfo AggregationInfoValues `json:"aggregation_info"`
}

type AggregationInfoValues struct {
	AggregationInfo map[string]float64 `json:"aggregation_info"` //ex: "avg": 123.340
	Format          string             `json:"format"`
}

func getResults(query string) float64 {
	var value float64
	//setup context with a timeout to avoid blocking
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, err := promApi.NewClient(promApi.Config{
		Address: config.Cfg.ThanosURL,
	})
	if err != nil {
		klog.Errorf("Error creating client: %v. Please ensure that the API server is running and the address is correct", err)
	}
	v1api := promv1.NewAPI(client)
	res, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			klog.Errorf("API query timed out: %v", err)
		}
		klog.Errorf("API query failed: %v", err)
	}

	vector := res.(model.Vector)

	for _, sample := range vector {
		value = (float64)(sample.Value)
	}

	return value
}
