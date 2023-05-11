package server

type ClusterNamespaceMap struct {
	ClusterNamespace map[string]string
	//map[string]in64 --> ex: {"ns_local-cluster_open-cluster-management-observability": 12343212}
}

type RecommendationIDMap struct {
	RecommendationID map[string]string
	//map[int64]string --> ex. {12343212: "ns_local-cluster_open-cluster-management-observability_deployment_container"
}

type RecommendationStatusMap struct {
	RecommendationStatus map[string]string
	// lock                 sync.Mutex
	// map[string]string --> ex. "ns_local-cluster_open-cluster-management-observability_12343212-deployment_container": "Good"
}
