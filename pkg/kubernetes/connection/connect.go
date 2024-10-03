package connection

import (
	"k8s-resource-autoscaler/pkg/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// ConnectToCluster connects to the Kubernetes cluster using the kubeconfig file
func ConnectToCluster() *kubernetes.Clientset {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Error("Error building kubeconfig: %v", err)
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error("Error creating Kubernetes clientset: %v", err)
		return nil
	}

	return clientset
}
