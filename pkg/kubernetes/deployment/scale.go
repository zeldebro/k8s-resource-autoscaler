package deployment

import (
	"context"
	"k8s-resource-autoscaler/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// ScalePod scales the deployment to the desired replica count with retry logic.
func ScalePod(clientset *kubernetes.Clientset, namespace string, n string, desiredReplicaCount int32) error {
	// Get all deployments in the namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal("Failed to list deployments: %v", err)
		return err
	}

	for _, deployment := range deployments.Items {
		deploymentName := deployment.Name

		// Check for scaling annotations
		annotations := deployment.Spec.Template.Annotations
		scaleUp := annotations["autoscale.k8s.io/scale-up"] == "true"
		scaleDown := annotations["autoscale.k8s.io/scale-down"] == "true"

		// Adjust the desiredReplicaCount based on ingress and egress
		if scaleUp {
			log.Info("Scaling up deployment %s based on ingress and egress data", deploymentName)
			desiredReplicaCount = *deployment.Spec.Replicas + 1
		} else if scaleDown {
			log.Info("Scaling down deployment %s based on ingress and egress data", deploymentName)
			if *deployment.Spec.Replicas > 1 {
				desiredReplicaCount = *deployment.Spec.Replicas - 1
			} else {
				desiredReplicaCount = 1 // Ensure at least one replica
			}
		}

		// Scale the deployment with retry mechanism
		maxRetries := 3                 // Maximum number of retries
		waitDuration := 2 * time.Second // Wait duration between retries
		for i := 0; i < maxRetries; i++ {
			scaleResponse, err := clientset.AppsV1().Deployments(namespace).GetScale(context.TODO(), deploymentName, metav1.GetOptions{})
			if err != nil {
				log.Fatal("Failed to get scale: %v", err)
				return err
			}

			scaleResponse.Spec.Replicas = desiredReplicaCount
			log.Info("Setting desired replicas for deployment %s to %v", deploymentName, desiredReplicaCount)

			_, err = clientset.AppsV1().Deployments(namespace).UpdateScale(context.TODO(), deploymentName, scaleResponse, metav1.UpdateOptions{})
			if err != nil {
				log.Error("Failed to update scale for deployment %s: %v", deploymentName, err)
				if i < maxRetries-1 { // Log and wait if not the last attempt
					log.Info("Retrying to scale deployment %s... (attempt %d)", deploymentName, i+2)
					time.Sleep(waitDuration)
				} else {
					return err // Return the last error after final attempt
				}
			} else {
				log.Info("Deployment %s scaled successfully to %v replicas", deploymentName, desiredReplicaCount)
				break // Exit loop on success
			}
		}
	}

	return nil
}

// WaitForScaling waits for the deployment to reach the desired replica count.
func WaitForScaling(clientset *kubernetes.Clientset, deploymentName, namespace string, desiredReplicaCount int32) error {
	timeout := time.After(1 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return log.Error("Timed out waiting for scaling of deployment %s to %d replicas", deploymentName, desiredReplicaCount)
		case <-ticker.C:
			deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
			if err != nil {
				log.Error("Failed to get deployment status: %v", err)
				return err
			}
			log.Info("Current replicas: %d, Desired replicas: %d", deployment.Status.Replicas, desiredReplicaCount)

			if deployment.Status.ReadyReplicas == desiredReplicaCount {
				log.Info("Deployment %s scaled successfully to %v replicas", deploymentName, desiredReplicaCount)
				return nil
			}
		}
	}
}

// ScaleDownOldPods gradually scales down the old pods in a deployment.
func ScaleDownOldPods(clientset *kubernetes.Clientset, deploymentName, namespace string) error {
	// Retrieve the current deployment
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return log.Error("Deployment %s not found in namespace %s", deploymentName, namespace)
		}
		return log.Error("Failed to get deployment %s in namespace %s: %v", deploymentName, namespace, err)
	}

	// Calculate the new replica count (scaling down by 1)
	if *deployment.Spec.Replicas > 1 {
		newReplicaCount := *deployment.Spec.Replicas - 1
		deployment.Spec.Replicas = &newReplicaCount

		// Update the deployment with the new replica count
		_, err = clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return log.Error("Failed to scale down deployment %s in namespace %s: %v", deploymentName, namespace, err)
		}
		log.Info("Deployment %s scaled down to %v replicas", deploymentName, newReplicaCount)
	} else {
		log.Info("Deployment %s is already at 1 replica, no further scaling down is possible", deploymentName)
	}

	return nil
}

// WaitForPVCReady waits for a Persistent Volume Claim (PVC) to be ready.
func WaitForPVCReady(clientset *kubernetes.Clientset, pvcName, namespace string) error {
	timeout := time.After(1 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return log.Error("Timed out waiting for PVC %s in namespace %s to be ready", pvcName, namespace)
		case <-ticker.C:
			pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
			if err != nil {
				log.Error("Failed to get PVC %s in namespace %s: %v", pvcName, namespace, err)
				return err
			}

			if pvc.Status.Phase == "Bound" {
				log.Info("PVC %s in namespace %s is ready", pvcName, namespace)
				return nil
			}
		}
	}
}
