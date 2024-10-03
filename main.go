package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s-resource-autoscaler/config"
	"k8s-resource-autoscaler/pkg/kubernetes/annotations"
	"k8s-resource-autoscaler/pkg/kubernetes/connection"
	"k8s-resource-autoscaler/pkg/kubernetes/deployment"
	"k8s-resource-autoscaler/pkg/kubernetes/metrics"
	"k8s-resource-autoscaler/pkg/kubernetes/pvc"
	"k8s-resource-autoscaler/pkg/log"
)

func main() {
	// Define a command-line flag for selecting the mode (pvc, ingress, or both)
	mode := flag.String("mode", "", "Mode of operation: 'pvc' for PVC resizing, 'ingress' for ingress scaling, 'pvc,ingress' for both")
	flag.Parse()

	// Initialize the logger
	log.Init()
	log.Info("Starting Kubernetes Resource Autoscaler...")

	// Ensure a valid mode is provided
	if *mode != "pvc" && *mode != "ingress" && *mode != "pvc,ingress" {
		fmt.Println("Error: You must specify a valid mode ('pvc', 'ingress', or 'pvc,ingress')")
		flag.Usage()
		os.Exit(1)
	}

	// Load the configuration from the YAML file
	configPath := "config.yaml"
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Error loading config: %v", err)
		os.Exit(1)
	}

	// Connect to the Kubernetes cluster
	clientset := connection.ConnectToCluster()

	// Continuous monitoring loop
	for {
		log.Info("Starting new monitoring cycle...")

		// Check for deployments with the specified annotation
		results, annotationFound, err := annotations.IsAnnotation(clientset)
		if err != nil {
			log.Error("Error checking annotations: %v", err)
			return
		}

		// Log results of annotation checks
		if annotationFound {
			log.Info("Found %d deployments with annotations.", len(results))
			for _, result := range results {
				log.Info("Deployment: %s, Namespace: %s, PVCs: %v", result.Deployment, result.Namespace, result.PVCNames)
			}
		} else {
			log.Error("No deployments with the specified annotation found.")
			return
		}

		// Determine which modes to run
		runPVC := strings.Contains(*mode, "pvc")
		runIngress := strings.Contains(*mode, "ingress")

		if runPVC {
			log.Info("Running in PVC resizing mode...")
			for _, result := range results {
				log.Info("Checking PVCs for deployment %s in namespace %s", result.Deployment, result.Namespace)

				for _, pvcName := range result.PVCNames {
					// Fetch disk usage percentage using PVC name and namespace
					diskUsagePercentage, err := metrics.FetchDiskUsage(config.Prometheus.URL, pvcName, result.Namespace)
					if err != nil {
						log.Error("Error fetching disk usage for PVC %s in namespace %s: %v", pvcName, result.Namespace, err)
						continue
					}

					log.Info("Disk usage for PVC %s in namespace %s: %.2f%%", pvcName, result.Namespace, diskUsagePercentage)

					// Check if disk usage exceeds threshold (convert to int for comparison)
					if int(diskUsagePercentage) > config.Thresholds.DiskUsage.Resize {
						err = pvc.ResizePVC(clientset, pvcName, result.Namespace)
						if err != nil {
							log.Error("Error resizing PVC %s in namespace %s: %v", pvcName, result.Namespace, err)
							continue
						}
						log.Info("Resized PVC %s in namespace %s successfully.", pvcName, result.Namespace)

						err = deployment.WaitForPVCReady(clientset, pvcName, result.Namespace)
						if err != nil {
							log.Error("Error waiting for PVC %s to be ready: %v", pvcName, result.Namespace, err)
							continue
						}
						log.Info("PVC %s is ready.", pvcName)
					} else {
						log.Info("Disk usage for PVC %s is below threshold, no resizing needed.", pvcName)
					}
				}
			}
		}

			if runIngress {
    			log.Info("Running in ingress scaling mode...")
    			for _, result := range results {
    				log.Info("Checking network usage for deployment %s in namespace %s", result.Deployment, result.Namespace)

    				// Get the list of pods for the deployment
    				pods, err := deployment.GetPodsForDeployment(clientset, result.Deployment, result.Namespace)
    				if err != nil {
    					log.Error("Error fetching pods for deployment %s in namespace %s: %v", result.Deployment, result.Namespace, err)
    					continue
    				}

    				// Iterate over each pod to fetch ingress and egress bandwidth
    				for _, pod := range pods {
    					log.Info("Checking network usage for pod %s in namespace %s", pod.Name, result.Namespace)

    					// Fetch ingress and egress bandwidth from Prometheus for the pod
    					ingressBandwidth, egressBandwidth, err := metrics.FetchNetworkUsage(config.Prometheus.URL, pod.Name, result.Namespace)
    					if err != nil {
    						log.Error("Error fetching network usage for pod %s in namespace %s: %v", pod.Name, result.Namespace, err)
    						continue
    					}

    					log.Info("\n Ingress Bandwidth: %.2f bytes/sec, Egress Bandwidth: %.2f bytes/sec for pod %s in namespace %s",
    						ingressBandwidth, egressBandwidth, pod.Name, result.Namespace)

    					// Check if ingress exceeds threshold (convert to int for comparison)
    					if int(ingressBandwidth) > config.Thresholds.NetworkUsage.Ingress.Scale {
    						log.Info("Ingress bandwidth for pod %s exceeds threshold. Scaling deployment %s in namespace %s to %d replicas...",
    							pod.Name, result.Deployment, result.Namespace, config.DesiredReplicaCount)

    						// Scale the deployment based on the network usage
    						err = deployment.ScalePod(clientset, result.Deployment, result.Namespace, int32(config.DesiredReplicaCount))
    						if err != nil {
    							log.Error("Error scaling deployment %s in namespace %s: %v", result.Deployment, result.Namespace, err)
    							continue
    						}
    					}
    				}
    			}
    		}
		// Wait for the specified interval before running the next cycle
		log.Info("Monitoring cycle complete. Waiting for %d minutes before the next cycle.", config.Interval)
		time.Sleep(time.Duration(config.Interval) * time.Minute)
	}
}
