package server

/*
Sample computeRecommendations POST request and response body: (NOTE: application is optional if namespace is provided)

https://localhost:4020/computeRecommendations

[
    {
        "clusterName": "local-cluster",
        "namespace": "open-cluster-management-observability",
        "application": "",
        "measurement_duration": "60mn"
    }
]

Sample computeRecommendations response body:

Recommendation for cluster local-cluster namespace open-cluster-management-observability
successfully submitted with recommendation Id ns_local-cluster_open-cluster-management-observability_00610619
*/

/*
Sample of getRecommendations POST request: (NOTE: either recommendation_id or combintation of cluster_name and namespace must be provided)

https://localhost:4020/getRecommendations

[
    {
        "recommendation_id": "",
        "cluster_name": "local-cluster",
        "namespace": "open-cluster-management-observability"

    }
]

Sample of getRecommendations response:
[
  {
    "cluster_name": "local-cluster",
    "kubernetes_objects": [
      {
        "type": "deployment",
        "name": "observability-thanos-rule",
        "namespace": "open-cluster-management-observability",
        "containers": [
          {
            "container_image_name": "configmap-reloader",
            "container_name": "configmap-reloader",
            "recommendations": {
              "notifications": [
                {
                  "type": "info",
                  "message": "There is not enough data available to generate a recommendation."
                }
              ],
              "data": {}
            }
          }
        ]
      }
    ],
    "version": "1.0",
    "experiment_name": "ns_local-cluster_open-cluster-management-observability_00610619-observability-thanos-rule-configmap-reloader"
  }
]

*/
