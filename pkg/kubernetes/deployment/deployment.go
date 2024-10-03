package deployment

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Pod struct {
	Name      string
	Namespace string
}

// GetPodsForDeployment retrieves the pods associated with a given deployment
func GetPodsForDeployment(clientset *kubernetes.Clientset, deploymentName, namespace string) ([]Pod, error) {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{}) // Use metav1.GetOptions
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}

	var podList []Pod
	for _, pod := range pods.Items {
		podList = append(podList, Pod{Name: pod.Name, Namespace: pod.Namespace})
	}

	return podList, nil
}
