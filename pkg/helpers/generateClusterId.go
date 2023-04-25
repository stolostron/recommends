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

//converts cpu usage to cores
func ConvertToCores(cpuUsage float64, coresAvailable float64) float64 {

	// coresAvailable := 4.0 //assume 4 cores available

	// cpuUsage := 0.5 //cpu usage is 0.5

	cpuUsageInCores := cpuUsage * coresAvailable

	return cpuUsageInCores

}
