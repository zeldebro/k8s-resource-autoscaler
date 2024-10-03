```markdown
# Kubernetes Resource Autoscaler

## Project Overview
The Kubernetes Resource Autoscaler automatically adjusts the size of Persistent Volume Claims (PVC) based on capacity and monitors ingress and egress bandwidth.

## Features
- Automatically scales PVC based on defined metrics.
- Monitors ingress and egress bandwidth.
- Easy integration with existing Kubernetes clusters.

## Prerequisites
- Go version **1.19** or higher
- Kubernetes cluster
- Prometheus (for monitoring metrics)

## Installation Instructions
1. Clone the repository:
   ```bash
   git clone https://github.com/zeldebro/k8s-resource-autoscaler.git
   cd k8s-resource-autoscaler
   ```

2. Ensure you have Go installed and set up:
   ```bash
   go version # Check if Go is installed
   ```

3. Run `go mod tidy` to install necessary dependencies:
   ```bash
   go mod tidy
   ```

4. Start the application in different modes:
   - To run the autoscaler for PVCs:
     ```bash
     go run main.go --mode=pvc
     ```
   - To run the autoscaler for both PVCs and ingress:
     ```bash
     go run main.go --mode=pvc,ingress
     ```
   - To run the autoscaler for ingress only:
     ```bash
     go run main.go --mode=ingress
     ```

## Usage Guidelines
- Configure the autoscaler by modifying the configuration files as needed.
- Ensure that Prometheus is correctly set up to gather metrics.

## Configuration Examples

### Scenario: Automatic PVC Resizing
Suppose you have a PVC named `foo-pvc` with an initial size of **1Gi**. As data is added, the PVC reaches its capacity, and you don't want to restart or kill the pod. By setting a threshold for the PVC, the autoscaler can automatically resize it when it reaches the specified limit.

1. **Initial PVC Configuration**:
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: foo-pvc
   spec:
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 1Gi
   ```

2. **Threshold Configuration**:
   In your `values.yaml`, you can define a threshold that triggers resizing:
   ```yaml
   pvc:
     name: foo-pvc
     threshold: 90 # percentage
     increment: 0.5 # increase size by 0.5Gi when threshold is reached
   ```

3. **Expected Behavior**:
   When the usage of `foo-pvc` reaches **90%**, the autoscaler will automatically resize the PVC from **1Gi** to **1.5Gi** without requiring any pod restarts or manual intervention.

### Ingress Example
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: myapp.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80
```

### Egress Example
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: my-egress-policy
spec:
  podSelector:
    matchLabels:
      app: my-app
  policyTypes:
  - Egress
  egress:
  - to:
    - podSelector:
        matchLabels:
          role: db
    ports:
    - protocol: TCP
      port: 5432
```

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [Prometheus Documentation](https://prometheus.io/docs/introduction/overview/)
