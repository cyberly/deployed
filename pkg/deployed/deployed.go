package deployed

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
)

var (
	httpClient                       = &http.Client{}
	k8sClient  *kubernetes.Clientset = newK8sClient()
)

//Bootstrap - Entry point
func Bootstrap() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/verify", verifyHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":80", router))
}

func isURLInvalid(u string) bool {
	v, err := url.Parse(u)
	return err != nil || v.Scheme == "" || v.Host == ""
}
