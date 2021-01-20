package deployed

import (
	"encoding/json"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

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

func watchDeploymentEvents(req verifyRequest) {
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
		reqStatus := notifyPipeline(req.getPipeline())
		log.Printf("[%v] Notified Azure Devops pipeline, got \"%v\"", *req.Namespace, reqStatus)
	case <-time.After(time.Duration(req.Timeout) * time.Second):
		log.Printf("[%v] Timeout exceeded looking for %v", req.Namespace, req.Image)
	}
	close(listenerCh)
}
