```markdown
# Kubernetes Resource Autoscaler

## Overview

This project is a Kubernetes Resource Autoscaler designed to automatically increase Persistent Volume Claim (PVC) size based on disk capacity and monitor ingress and egress traffic for deployments. It uses Prometheus to collect metrics and trigger scaling actions based on defined thresholds.

## Features

- Automatic resizing of PVCs when disk usage exceeds defined thresholds.
- Monitoring of ingress and egress network traffic for deployments.
- Dynamic scaling of deployments based on network usage metrics.
- Customizable configurations through a YAML file.

## Prerequisites

- Kubernetes cluster with sufficient permissions to manage PVCs and deployments.
- Prometheus installed and configured to scrape metrics from the Kubernetes cluster.
- Go (version 1.18 or higher) for building the project.

## Getting Started

1. **Clone the Repository:**

   ```bash
   git clone https://your-repo-url.git
   cd k8s-resource-autoscaler
   ```

2. **Install Dependencies:**

   Ensure you have Go modules enabled and install necessary dependencies.

   ```bash
   go mod tidy
   ```

3. **Configure Storage Class Annotations:**

   To enable the autoscaling feature for PVCs, you need to ensure that your storage class has the correct annotations. Here's an example of a storage class with the required annotations:

   ```yaml
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: your-storage-class
   provisioner: your-provisioner
   parameters:
     allowVolumeExpansion: "true"  # This allows PVC size to be increased.
   ```

4. **Prometheus Configuration:**

   Ensure Prometheus is set up to scrape metrics from the necessary endpoints. Your Prometheus configuration should include the following job definitions to scrape Kubernetes metrics:

   ```yaml
   scrape_configs:
     - job_name: 'kubernetes-pods'
       kubernetes_sd_configs:
         - role: pod
       relabel_configs:
         - source_labels: [__meta_kubernetes_namespace]
           action: keep
           regex: your-namespace
   ```

5. **Configuration File:**

   Create a configuration file `config.yaml` with the following structure:

   ```yaml
   Prometheus:
     URL: "http://your-prometheus-url"
   Thresholds:
     DiskUsage:
       Resize: 80  # Resize threshold percentage for PVCs
     NetworkUsage:
       Ingress:
         Scale: 1000000  # Ingress bandwidth threshold in bytes/sec
   DesiredReplicaCount: 3  # Desired number of replicas for scaling
   Interval: 5  # Monitoring interval in minutes
   ```

6. **Running the Autoscaler:**

   You can run the autoscaler with the following command, specifying the mode of operation (either `pvc`, `ingress`, or both):

   ```bash
   go run main.go -mode pvc,ingress
   ```

## Usage

- The autoscaler continuously monitors PVC usage and network traffic.
- When disk usage exceeds the configured threshold, it automatically resizes the PVC.
- If ingress traffic exceeds the defined threshold, the autoscaler scales up the deployment to handle increased traffic.

## Logs

The application logs can provide insights into the operations performed by the autoscaler. Logs will indicate when PVCs are resized and when deployments are scaled based on network usage.

## Example Metrics

- **PVC Resizing**: If a PVC usage exceeds 80%, the autoscaler will attempt to resize it automatically.
- **Ingress Bandwidth**: If ingress bandwidth exceeds 1,000,000 bytes/sec, the autoscaler will scale the deployment up to the desired number of replicas.

## Contributing

Contributions are welcome! If you have suggestions or improvements, feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License.

```

### Notes
- **Adjust Links**: Replace `https://your-repo-url.git` and `http://your-prometheus-url` with actual URLs for your repository and Prometheus server.
- **Custom Parameters**: Modify any specific parameters as needed based on your implementation details.
- **Annotations**: Ensure the annotations mentioned are applicable to your storage class and required for your PVCs.
