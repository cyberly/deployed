package deployed

import (
	"encoding/json"
	"log"
	"net/http"
)

func appConfigHandler(w http.ResponseWriter, r *http.Request) {
	var secretData map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&secretData)
	if err != nil {
		log.Panicf("Error unmarshalling secret data: %v ", err.Error())
	}

	pipeline := getPipelineFromHeaders(r)
	namespace := r.Header.Get("namespace")
	secret := r.Header.Get("secretName")

	log.Printf("Namesapce: %v, Secret Name: %v, PlanID: %v", namespace, secret, pipeline.PlanURL)
	log.Println(secretData)
}
