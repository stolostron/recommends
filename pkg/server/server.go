package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/stolostron/recommends/pkg/config"

	klog "k8s.io/klog/v2"

	"github.com/gorilla/mux"
)

func StartAndListen() {
	port := config.Cfg.HttpPort

	// Configure TLS
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	// Initiate router
	router := mux.NewRouter()
	router.HandleFunc("/liveness", livenessProbe).Methods("GET")
	router.HandleFunc("/readiness", readinessProbe).Methods("GET")
	router.HandleFunc("/computeRecommendation", computeRecommendations).Methods("POST")
	router.HandleFunc("/listRecommendation", listRecommendations).Methods("POST")

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		TLSConfig:         cfg,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	klog.Infof(`Recommends API is now running on https://localhost:%d`, port)
	serverErr := srv.ListenAndServeTLS("./sslcert/tls.crt", "./sslcert/tls.key")
	if serverErr != nil {
		klog.Fatal("Server process ended with error. ", serverErr)
	}
}
