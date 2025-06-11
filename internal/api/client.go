package api

import (
	"context"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"
)

func NewClient() (*kubernetes.Clientset, error) {
	//configpath := "C:\\Users\\侯哥哥\\.kube\\config"
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	//confignew, err := clientcmd.BuildConfigFromFlags("", configpath)
	//if err != nil {
	//	inClusterConfig, err := rest.InClusterConfig()
	//	if err != nil {
	//		log.Fatalln("can't find config")
	//	}
	//	confignew = inClusterConfig
	//}
	//2.creat clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln("can't create clientset")
	}
	context.WithTimeout(context.Background(), 10*time.Second)
	_, err = clientset.CoreV1().Namespaces().List(context.Background(), metaV1.ListOptions{})
	if err != nil {
		log.Fatalln("can't create clientset")
	}
	return clientset, nil
}
