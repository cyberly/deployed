package deployed

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

type pipeline struct {
	PlanURL        string `json:"PlanUrl"`
	ProjectID      string `json:"ProjectId"`
	HubName        string `json:"HubName"`
	PlanID         string `json:"PlanId"`
	JobID          string `json:"JobId"`
	TimelineID     string `json:"TimelineId"`
	TaskInstanceID string `json:"TaskInstanceId"`
	AuthToken      string `json:"AuthToken"`
}

func buildPipelinePayload(p pipeline) []byte {
	payload := map[string]string{
		"Name":   "TaskCompleted",
		"TaskID": p.TaskInstanceID,
		"JobID":  p.JobID,
		"Result": "successed",
	}
	reqPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to build pipeline poayload: %v", err.Error())
	}
	return reqPayload
}

func buildPipelineRequest(p pipeline) *http.Request {
	req, err := http.NewRequest("POST",
		buildPipelineURL(p),
		bytes.NewBuffer(buildPipelinePayload(p)))
	if err != nil {
		log.Printf("Failed assembling pipeline request: %v", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(":"+p.AuthToken)))
	return req
}

func buildPipelineURL(p pipeline) string {
	return p.PlanURL + p.ProjectID + "/_apis/distributedtask/hubs/" +
		p.HubName +
		"/plans/" +
		p.PlanID +
		"/events?api-version=2.0-preview.1"
}

func getPipelineFromHeaders(req *http.Request) pipeline {
	p := &pipeline{
		PlanURL:        req.Header.Get("PlanUrl"),
		ProjectID:      req.Header.Get("ProjectId"),
		HubName:        req.Header.Get("HubName"),
		PlanID:         req.Header.Get("PlanId"),
		JobID:          req.Header.Get("JobId"),
		TimelineID:     req.Header.Get("TimelineId"),
		TaskInstanceID: req.Header.Get("TaskInstanceId"),
		AuthToken:      req.Header.Get("AuthToken"),
	}
	return *p
}

func notifyPipeline(p pipeline) string {
	req := buildPipelineRequest(p)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error talking to Azure Devops: %v", err.Error())
	}
	defer resp.Body.Close()
	return resp.Status
}
