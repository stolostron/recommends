package server

// example:
// 		{ "deployment1" : [
// 				{"container1" : "recommendation1"},
// 				{"container2" :"recommendation2"},..],
// 		{ "deployment2" : [
// 				{"container1" : "recommendation1"},
// 				{"container2" :"recommendation2"},..]
// 		}
// 		}

type RecommendationItem struct {
	Cluster              string
	Namespace            string
	RecommendationID     string
	RecommendationStatus string
	Recommendation       map[string][]map[string]string
}
