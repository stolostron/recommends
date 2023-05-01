package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"k8s.io/klog/v2"
)

func GenerateID() string {

	max := big.NewInt(999099)
	valBig, err := rand.Int(rand.Reader, max)
	if err != nil {
		klog.Warning("Error generating RequestID.")
		return fmt.Sprint(0)
	}
	return fmt.Sprint(valBig.Int64())
}

func RemoveDuplicate(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func ConvertCpuUsageToCores(cpuUsage float64) float64 {
	// Convert CPU usage from millicores to cores
	cores := cpuUsage * 1000.0
	return cores
}

func ConvertMemoryUsageToMiB(memoryUsage float64) float64 {
	// Convert memory usage from bytes to Mebibytes (MiB)
	miB := memoryUsage / 1024.0 / 1024.0
	return miB
}
