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
	Value           *float64              `json:"value,omitempty"`
	Format          string                `json:"format,omitempty"`
	AggregationInfo AggregationInfoStruct `json:"aggregation_info"`
}
type AggregationInfoStruct struct {
	Avg    *float64 `json:"avg"`
	Max    *float64 `json:"max,omitempty"`
	Min    *float64 `json:"min,omitempty"`
	Sum    *float64 `json:"sum"`
	Format string   `json:"format"`
}

///takes in queries from the performance profile and queries thanos then dumps results in updateResults
//get the query name, function and query per workload get metrics, return metrics

func GetMetricsForQuery(queryNameMap map[string][]string) *Metrics {
	var MetricsList []Metrics
	var metrics Metrics
	var format string
	var aggregateStruct AggregationInfoStruct
	var resultValue *float64

	for name, queries := range queryNameMap {
		if strings.Contains(name, "cpu") {
			format = "cores"

		} else {
			format = "MiB"
		}
		for _, query := range queries {
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
			var avg, sum, max, min *float64
			for _, sample := range vector {
				klog.Infof("Name: %s, Value: %v, Time: %d, Query: %s", name, sample.Value, sample.Timestamp, query)

				function := strings.Split(query, "(")[0] //get function

				if function == "avg" {
					avg = (*float64)(&sample.Value)
					resultValue = (*float64)(&sample.Value) //resultValue == avg
				}
				if function == "min" {
					min = (*float64)(&sample.Value)
				}
				if function == "max" {
					max = (*float64)(&sample.Value)
				}
				if function == "sum" {
					sum = (*float64)(&sample.Value)
				}

				aggregateStruct = AggregationInfoStruct{
					Avg:    avg,
					Max:    max,
					Min:    min,
					Sum:    sum,
					Format: format,
				}
			}
		}

		metrics.Name = name
		metrics.Results = Result{Value: resultValue,
			Format: format, AggregationInfo: aggregateStruct,
		}

		klog.Info(metrics)
	}

	MetricsList = append(MetricsList, metrics)

	return &metrics
}
