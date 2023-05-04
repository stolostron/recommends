package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"k8s.io/klog"
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

func ConvertCpuUsageToCores(millicpu float64) float64 {
	// Convert CPU usage from millicores to cores
	return millicpu * 1000.0
}

func ConvertMemoryUsageToMiB(bytes float64) float64 {
	// Convert memory usage from bytes to Mebibytes (MiB)
	return bytes / 1024.0 / 1024.0
}

//error handling for decoding request body:
func ErrorHandlingRequests(w http.ResponseWriter, err error) bool {
	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		var syntaxError *json.SyntaxError

		switch {

		case errors.As(err, &syntaxError):
			http.Error(w, "{\"message\":\"Request body contains badly-formed JSON.\"}", http.StatusBadRequest)

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %s field", unmarshalTypeError.Field)
			http.Error(w, msg, http.StatusBadRequest)

		case errors.Is(err, io.EOF):
			http.Error(w, "{\"message\":\"Request body must not be empty.\"}", http.StatusBadRequest)

		default:
			klog.Error(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return false
	}
	return true
}
