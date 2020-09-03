package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var (
	httpClient                       = &http.Client{}
	k8sClient  *kubernetes.Clientset = newK8sClient()
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

func buildPipelinePayload(v reqBody) []byte {
	payload := map[string]string{
		"Name":   "TaskCompleted",
		"TaskID": *v.TaskInstanceID,
		"JobID":  *v.JobID,
		"Result": "successed",
	}
	p, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to build pipeline poayload: %v", err.Error())
	}
	return p
}

func buildPipelineURL(v reqBody) string {
	return *v.PlanURL + *v.ProjectID + "/_apis/distributedtask/hubs/" +
		*v.HubName +
		"/plans/" +
		*v.PlanID +
		"/events?api-version=2.0-preview.1"
}

func checkImage(i string, containers []corev1.Container) bool {
	for _, c := range containers {
		if c.Image == i {
			return true
		}
	}
	return false

}

func checkConditions(conditions []appsv1.DeploymentCondition) bool {
	for _, c := range conditions {
		if c.Type != "Progressing" {
			continue
		}
		if c.Status != "True" {
			continue
		}
		if c.Reason != "NewReplicaSetAvailable" {
			continue
		}
		return true
	}
	return false
}

func checkDeployStatus(e interface{}, ch chan bool, image string) {
	d := convertEvent(e)
	if !checkImage(image, d.Spec.Template.Spec.Containers) {
		return
	}
	if !checkGeneration(d) {
		return
	}
	if !checkConditions(d.Status.Conditions) {
		return
	}
	if !checkReplicas(d) {
		return
	}
	log.Printf("[%v] %v deployed", d.ObjectMeta.Namespace, d.ObjectMeta.Name)
	ch <- true
}

func checkGeneration(d appsv1.Deployment) bool {
	return d.ObjectMeta.Generation == d.Status.ObservedGeneration
}

func checkReplicas(d appsv1.Deployment) bool {
	if d.Status.UpdatedReplicas != *d.Spec.Replicas {
		return false
	}
	if d.Status.ReadyReplicas != *d.Spec.Replicas {
		return false
	}
	return true
}

func convertEvent(o interface{}) appsv1.Deployment {
	var d appsv1.Deployment
	j, _ := json.Marshal(o)
	json.Unmarshal(j, &d)
	return d
}

func newK8sClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

func notifyPipeline(v reqBody) {
	req, err := http.NewRequest("POST", buildPipelineURL(v), bytes.NewBuffer(buildPipelinePayload(v)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+*v.AuthToken)))
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err.Error())
	}
	defer resp.Body.Close()
	log.Printf("[%v] Notified Azure Devops pipeline, got \"%v\"", *v.Namespace, resp.Status)

}

func watchDeploymentEvents(req reqBody) {
	listenerCh := make(chan struct{})
	statusCh := make(chan bool)
	// Configure deployment listener
	watchlist := cache.NewListWatchFromClient(
		k8sClient.AppsV1().RESTClient(),
		"deployments",
		*req.Namespace,
		fields.Everything())
	// Start the listener and specify handlers for event types
	_, controller := cache.NewInformer(
		watchlist,
		&appsv1.Deployment{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				checkDeployStatus(obj, statusCh, *req.Image)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				checkDeployStatus(newObj, statusCh, *req.Image)
			},
		},
	)
	go controller.Run(listenerCh)
	select {
	case <-statusCh:
		notifyPipeline(req)
	case <-time.After(time.Duration(req.Timeout) * time.Second):
		log.Printf("[%v] Timeout exceeded looking for %v", req.Namespace, req.Image)
	}
	close(listenerCh)
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
	*req.Image = strings.ToLower(*req.Image)
	log.Printf("[%v] Watching for %v (timeout %v)", *req.Namespace, *req.Image, req.Timeout)
	go watchDeploymentEvents(req)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/verify", verifyHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":80", router))
}
