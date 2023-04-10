package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/golang/gddo/httputil/header"
	"github.com/stolostron/recommends/pkg/helpers"
	"k8s.io/klog"
)

type recommendation []struct {
	ClusterName         string `json:"clusterName"`
	Namespace           string `json:"namespace"`
	Application         string `json:"application"`
	MeasurementDuration string `json:"measurement_duration"` //ex: "15min"
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

	//get the clusterID (cluster name with namespace or applicaiton):
	clusterID := make(map[string]string) //ex: clustername-namespace:"id-12345"
	var concat string

	//use namespace:
	if newRecommendation[0].Application == "" && newRecommendation[0].Namespace != "" {
		concat = fmt.Sprintf("%s-%s", newRecommendation[0].ClusterName, newRecommendation[0].Namespace)
	}
	//use application
	if newRecommendation[0].Application != "" && newRecommendation[0].Namespace == "" {
		concat = fmt.Sprintf("%s-%s", newRecommendation[0].ClusterName, newRecommendation[0].Application)

		// if both applications and namespace is empty return
	} else if newRecommendation[0].Application == "" && newRecommendation[0].Namespace == "" {
		klog.V(4).Info("Request missing both Application and Namespace. Need at least one to fulfill request.")
		http.Error(w, "{\"message\":\"Both Application and Namespace cannot be empty.\"}", http.StatusBadRequest)
		return
	}

	//if id for clusterName already exists then we don't need to generate new oneß
	if clusterID[concat] == "" {
		clusterID[concat] = helpers.GenerateID(clusterID)

	}
	//TODO: decide if we need this
	//append to recommendations list temporary store in memory
	//recommendations = append(recommendations, newRecommendation...)

	msg := fmt.Sprintf("Recommendation for clusterID %s successfully submitted.", clusterID)
	_, err = w.Write([]byte(msg))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	klog.V(4).Info("Received recommendation request")
}
