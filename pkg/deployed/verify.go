package deployed

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type reqBody struct {
	Namespace      *string `json:"Namespace"`
	Image          *string `json:"Image"`
	Timeout        int     `json:"Timeout"`
	PlanURL        *string `json:"PlanUrl"`
	ProjectID      *string `json:"ProjectId"`
	HubName        *string `json:"HubName"`
	PlanID         *string `json:"PlanId"`
	JobID          *string `json:"JobId"`
	TimelineID     *string `json:"TimelineId"`
	TaskInstanceID *string `json:"TaskInstanceId"`
	AuthToken      *string `json:"AuthToken"`
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	var req = reqBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Unable to parse request body: %v", err.Error())
		return
	}
	if req.Timeout == 0 {
		req.Timeout = 180
	}
	if req.Namespace == nil {
		log.Println("Request body must contain namespace attribute")
		return
	}
	if req.Image == nil {
		log.Println("Request body must contain image attribute")
		return
	}
	if isURLInvalid(*req.PlanURL) {
		log.Println("Invalid plan URL.")
		return
	}
	*req.Image = strings.ToLower(*req.Image)
	log.Printf("[%v] Watching for %v (timeout %v)", *req.Namespace, *req.Image, req.Timeout)
	go watchDeploymentEvents(req)
}
