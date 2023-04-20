package kruize

import (
	"context"
	"errors"
	"strings"
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
	Value           string          `json:"value"`
	Format          string          `json:"format"`
	AggregationInfo AggregationInfo `json:"aggregation_info"`
}

type AggregationInfo struct {
	Function string `json:"function"`
	Query    string `json:"query"`
}

///takes in queries from the performance profile and queries thanos then dumps results in updateResults

//get the query name, function and query per workload get metrics, return metrics

func GetMetricsForQuery(queryNameMap map[string]string) Metrics {

	var metrics Metrics
	var format string

	for query, name := range queryNameMap {

		// klog.Infof("inside getMetricsForQuery for query: %s", query)
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
			klog.V(5).Infof("Name: %s, Value: %v, Time: %d, Metric: %s", name, sample.Value, sample.Timestamp, query)

			if strings.Contains(name, "cpu") {
				format = "cores"

			} else {
				format = "MiB"
			}

			metrics.Name = name
			metrics.Results = Result{Value: sample.Value.String(),
				Format: format, AggregationInfo: AggregationInfo{Function: strings.Split(query, "(")[0], Query: query}}

		}
	}

	return metrics
}
