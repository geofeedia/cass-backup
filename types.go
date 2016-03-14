package main

type CommonMetadata struct {
    cloud       string `json:"cloud"`
    region      string `json:"region"`
    zone        string `json:"zone"`
    hostname    string `json:"hostname"`
    instance_id string `json:"instanceId"`
    pod_name    string `json:"podName"`
}