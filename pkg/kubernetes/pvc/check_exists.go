package pvc

import (
	"context"
	"k8s-resource-autoscaler/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CheckPVCExists checks if a PVC exists in a specific namespace
func CheckPVCExists(clientset *kubernetes.Clientset, pvcName, namespace string) (bool, error) {
	log.Info("Checking if PVC %s exists in namespace %s", pvcName, namespace)

	_, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("PVC %s not found in namespace %s", pvcName, namespace)
			return false, nil
		}
		log.Error("Error checking PVC %s in namespace %s: %v", pvcName, namespace, err)
		return false, err
	}

	log.Info("PVC %s exists in namespace %s", pvcName, namespace)
	return true, nil
}
