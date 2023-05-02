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

// Convert CPU usage from millicores to cores
func ConvertCpuUsageToCores(millicpu float64) float64 {
	return millicpu * 1000.0
}

// Convert memory usage from bytes to Mebibytes (MiB)
func ConvertMemoryUsageToMiB(bytes float64) float64 {
	return bytes / 1024.0 / 1024.0
}
