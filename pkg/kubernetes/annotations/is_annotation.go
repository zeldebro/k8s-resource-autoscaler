package annotations

import (
	"context"
	"k8s-resource-autoscaler/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentResult struct {
	Namespace  string
	Deployment string
	PVCNames   []string
}

// IsAnnotation checks for deployments with the specified annotation in all namespaces.
func IsAnnotation(clientset *kubernetes.Clientset) ([]DeploymentResult, bool, error) {
	var results []DeploymentResult
	const annotationKey = "autoscaler/enabled"
	const annotationValue = "true"
	log.Info("Checking annotations......")
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, false, err
	}

	for _, ns := range namespaces.Items {
		deployments, err := clientset.AppsV1().Deployments(ns.Name).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, false, err
		}

		for _, deployment := range deployments.Items {
			if val, exists := deployment.Annotations[annotationKey]; exists && val == annotationValue {
				// Collect the PVCs used by the deployment, if any
				pvcNames := []string{}
				// Assuming your deployment spec has a template with volume claims
				for _, volume := range deployment.Spec.Template.Spec.Volumes {
					if volume.PersistentVolumeClaim != nil {
						pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
					}
				}

				results = append(results, DeploymentResult{
					Namespace:  ns.Name,
					Deployment: deployment.Name,
					PVCNames:   pvcNames,
				})
			}
		}
	}

	if len(results) == 0 {
		return nil, false, nil
	}
	return results, true, nil
}
