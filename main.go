package main

import (
	"flag"

	"github.com/stolostron/recommends/pkg/config"
	"github.com/stolostron/recommends/pkg/server"

	klog "k8s.io/klog/v2"
)

func main() {
	// Initialize the logger.
	klog.InitFlags(nil)
	flag.Parse()
	defer klog.Flush()
	klog.Info("Starting recommends.")

	// Read the config from the environment.
	config.Cfg.PrintConfig()

	// Validate required configuration to proceed.
	configError := config.Cfg.Validate()
	if configError != nil {
		klog.Fatal(configError)
	}

	// Load PerformanceProfile in Kruize Instance first
	// perfProfileInitialized := false
	// for !perfProfileInitialized {
	// 	klog.Info("Initializing performanceProfile.")
	// 	perfProfileInitialized = kruize.InitPerformanceProfile()
	// 	if perfProfileInitialized {
	// 		klog.Info("Initialized performanceProfile.")
	// 		break
	// 	} else {
	// 		klog.Info("Retry performanceProfile Initializing.")
	// 		klog.V(9).Infof("May be kruize is taking long to start ... Retry after 1 second")
	// 		time.Sleep(1 * time.Second)
	// 	}

	// }
	// server.GetLabels()
	server.StartAndListen()
}
