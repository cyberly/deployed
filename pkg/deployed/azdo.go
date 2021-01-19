package deployed

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

func buildPipelinePayload(r reqBody) []byte {
	payload := map[string]string{
		"Name":   "TaskCompleted",
		"TaskID": *r.TaskInstanceID,
		"JobID":  *r.JobID,
		"Result": "successed",
	}
	p, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to build pipeline poayload: %v", err.Error())
	}
	return p
}

func buildPipelineRequest(r reqBody) *http.Request {
	req, err := http.NewRequest("POST",
		buildPipelineURL(r),
		bytes.NewBuffer(buildPipelinePayload(r)))
	if err != nil {
		log.Printf("Failed assembling pipeline request: %v", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization",
		"Basic "+base64.StdEncoding.EncodeToString([]byte(":"+*r.AuthToken)))
	return req
}

func buildPipelineURL(r reqBody) string {
	return *r.PlanURL + *r.ProjectID + "/_apis/distributedtask/hubs/" +
		*r.HubName +
		"/plans/" +
		*r.PlanID +
		"/events?api-version=2.0-preview.1"
}

func notifyPipeline(r reqBody) {
	req := buildPipelineRequest(r)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error talking to Azure Devops: %v", err.Error())
	}
	defer resp.Body.Close()
	log.Printf("[%v] Notified Azure Devops pipeline, got \"%v\"", *r.Namespace, resp.Status)
}
