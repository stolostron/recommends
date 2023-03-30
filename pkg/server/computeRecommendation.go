package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/google/uuid"
	"k8s.io/klog"
)

type recommendation []struct {
	ClusterName         string `json:"clusterName"`
	Namespace           string `json:"namespace"`
	Application         string `json:"application"`
	MeasurementDuration string `json:"measurement_duration"` //ex: "15min"
	// ID                  string
}

var recommendations = recommendation{
	{ClusterName: "test-cluster", Namespace: "test-namespace", Application: "test-application", MeasurementDuration: "10min"},
}

// adds an recommendation from JSON received in the request body.
func computeRecommendations(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var newRecommendation recommendation

	err := dec.Decode(&newRecommendation)

	//error handling:
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
			log.Print(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if newRecommendation[0].Application == "" && newRecommendation[0].Namespace == "" {
		klog.V(4).Info("Request missing both Application and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both Application and Namespace cannot be empty.\"}", http.StatusBadRequest)
	}

	//concatenate unique id for incoming post clustername value
	uid := uuid.New()
	clusterName := fmt.Sprintf("%s-%s", newRecommendation[0].ClusterName, uid.String())
	newRecommendation[0].ClusterName = clusterName

	//append to recommendations list
	recommendations = append(recommendations, newRecommendation...)

	klog.Info("Received recommendation request")
	klog.V(4).Info(recommendations)
}
