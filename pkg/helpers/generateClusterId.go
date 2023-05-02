package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateID(object interface{}) string {

	objectString := fmt.Sprintf("%v", object)
	hash := sha256.Sum256([]byte(objectString))

	hexString := hex.EncodeToString(hash[:])

	return hexString
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
