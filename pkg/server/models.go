package server

type RecommendationData struct {
	ClusterNamespace     map[string]string //ex: {"ns_local-cluster_ocm": 12343212}
	RecommendationID     map[string]string // ex. {12343212: "ns_local-cluster_ocm" }
	RecommendationStatus map[string]string
	Recommendation       map[string][]map[string][]map[string]string
}

// example:
// {12343212:
// 	[
// 		{ "deployment1" : [
// 				{"container1" : "recommendation1"},
// 				{"container2" :"recommendation2"},..],
// 		{ "deployment2" : [
// 				{"container1" : "recommendation1"},
// 				{"container2" :"recommendation2"},..]
// 		}
// 		}
// 	]
// }

// map[string][]map[string][]map[string]string
