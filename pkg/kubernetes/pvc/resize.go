package pvc

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResizePVC resizes the PVC by 50% if its usage exceeds 50%.

func ResizePVC(clientset *kubernetes.Clientset, pvcName, namespace string) error {
	// Fetch the existing PVC
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metaV1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("PVC %s not found in namespace %s", pvcName, namespace)
		}
		return fmt.Errorf("error getting PVC %s: %v", pvcName, err)
	}

	// Get current PVC size
	currentSize := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	currentSizeValue := currentSize.Value()

	// Calculate the new size (increase by 50%)
	newSize := currentSizeValue + (currentSizeValue / 2)

	// Update the PVC size
	pvc.Spec.Resources.Requests[v1.ResourceStorage] = *resource.NewQuantity(newSize, resource.DecimalSI)

	// Attempt to update the PVC
	_, err = clientset.CoreV1().PersistentVolumeClaims(namespace).Update(context.TODO(), pvc, metaV1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("error updating PVC %s: %v", pvcName, err)
	}

	return nil
}
