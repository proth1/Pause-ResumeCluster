package main

type Secret struct {
	CastAIKey string `json:"CastAIKey"`
}

type ClusterItems struct {
	Items []ClusterInfo `json:"items"`
}

type ClusterInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ClustersToAction struct {
	Action       string   `json:"action"`
	ClusterNames []string `json:"clusterNames"`
}
