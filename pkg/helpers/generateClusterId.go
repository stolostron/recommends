package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"crypto/rand"
	"math/big"

	"k8s.io/klog/v2"
)

func GenerateID() string {

	max := big.NewInt(999099)
	valBig, err := rand.Int(rand.Reader, max)
	if err != nil {
		klog.Warningf("Error generating RequestID. %s", err.Error())
		return "0"
	}
	return fmt.Sprintf("%08d", valBig.Int64())
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
			klog.Error("Unexpected error while processing request. ", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return false
	}
	return true
}
