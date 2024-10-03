package metrics

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"gopkg.in/yaml.v2"
	"k8s-resource-autoscaler/pkg/log"
)

// PrometheusResponse represents the structure of the Prometheus query response
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Config represents the structure of the configuration file
type Config struct {
	Prometheus struct {
		DiskUsageQuery    string `yaml:"disk_usage_query"`
		NetworkUsageQueries struct {
			Ingress string `yaml:"ingress"`
			Egress  string `yaml:"egress"`
		} `yaml:"network_usage_queries"`
	} `yaml:"prometheus"`
}

// LoadConfig reads the configuration from a YAML file
func LoadConfig(filePath string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("error reading YAML file: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	return config, nil
}

// ReplacePlaceholders replaces placeholders in the query with actual values
func ReplacePlaceholders(query string, pvcName, namespace, podName string) string {
	query = strings.ReplaceAll(query, "{{pvc_name}}", pvcName)
	query = strings.ReplaceAll(query, "{{namespace}}", namespace)
	query = strings.ReplaceAll(query, "{{pod_name}}", podName)
	return query
}

// FetchDiskUsage queries Prometheus for disk usage percentage
func FetchDiskUsage(prometheusURL, pvcName, namespace string) (float64, error) {
	// Fetch config to get the query from the YAML file
	config, err := LoadConfig("config.yaml")
	if err != nil {
		return 0, err
	}

	query := ReplacePlaceholders(config.Prometheus.DiskUsageQuery, pvcName, namespace, "")

	// Encode the query for use in a URL
	encodedQuery := url.QueryEscape(query)

	// Construct the full URL for the Prometheus query
	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)

	// Log the full URL for debugging
	log.Info("Fetching disk usage from Prometheus at URL: %s", fullURL)

	// Make the HTTP GET request to Prometheus
	resp, err := http.Get(fullURL)
	if err != nil {
		log.Error("Error querying Prometheus: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error reading response from Prometheus: %v", err)
		return 0, err
	}

	// Unmarshal the response into the PrometheusResponse struct
	var prometheusResponse PrometheusResponse
	if err := json.Unmarshal(body, &prometheusResponse); err != nil {
		log.Error("Error unmarshalling Prometheus response: %v", err)
		return 0, err
	}

	// Check for a successful response and results
	if prometheusResponse.Status != "success" || len(prometheusResponse.Data.Result) == 0 {
		return 0, fmt.Errorf("no data returned from Prometheus for query: %s", query)
	}

	// Extract the disk usage value safely
	if len(prometheusResponse.Data.Result) > 0 {
		if value, ok := prometheusResponse.Data.Result[0].Value[1].(string); ok {
			var diskUsagePercentage float64
			if _, err := fmt.Sscanf(value, "%f", &diskUsagePercentage); err != nil {
				log.Error("Error parsing disk usage percentage: %v", err)
				return 0, err
			}
			return diskUsagePercentage, nil
		}
	}

	return 0, fmt.Errorf("unexpected data format in Prometheus response")
}

// FetchNetworkUsage queries Prometheus for ingress and egress network usage
func FetchNetworkUsage(prometheusURL, podName, namespace string) (float64, float64, error) {
	// Fetch config to get the queries from the YAML file
	config, err := LoadConfig("config.yaml") // Adjust path as necessary
	if err != nil {
		return 0, 0, err
	}

	// Use the queries from the config file
	ingressQuery := ReplacePlaceholders(config.Prometheus.NetworkUsageQueries.Ingress, "", namespace, podName)
	egressQuery := ReplacePlaceholders(config.Prometheus.NetworkUsageQueries.Egress, "", namespace, podName)

	// Encode the queries for use in a URL
	encodedIngressQuery := url.QueryEscape(ingressQuery)
	encodedEgressQuery := url.QueryEscape(egressQuery)

	// Construct the full URLs for the Prometheus queries
	ingressURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedIngressQuery)
	egressURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedEgressQuery)

	// Log the full URLs for debugging
	log.Info("\nFetching ingress network usage from Prometheus at URL: %s", ingressURL)
	log.Info("\nFetching egress network usage from Prometheus at URL: %s", egressURL)

	// Fetch ingress usage
	ingressResp, err := http.Get(ingressURL)
	if err != nil {
		log.Error("Error querying Prometheus for ingress: %v", err)
		return 0, 0, err
	}
	defer ingressResp.Body.Close()

	// Read the ingress response body
	ingressBody, err := ioutil.ReadAll(ingressResp.Body)
	if err != nil {
		log.Error("Error reading ingress response from Prometheus: %v", err)
		return 0, 0, err
	}

	// Unmarshal the ingress response into the PrometheusResponse struct
	var ingressResponse PrometheusResponse
	if err := json.Unmarshal(ingressBody, &ingressResponse); err != nil {
		log.Error("Error unmarshalling ingress Prometheus response: %v", err)
		return 0, 0, err
	}

	// Extract ingress value
	var ingress float64
	if ingressResponse.Status == "success" && len(ingressResponse.Data.Result) > 0 {
		if ingressValue, ok := ingressResponse.Data.Result[0].Value[1].(string); ok {
			if _, err := fmt.Sscanf(ingressValue, "%f", &ingress); err != nil {
				log.Error("Error parsing ingress network usage: %v", err)
				return 0, 0, err
			}
		}
	} else {
		log.Error("No successful ingress data returned from Prometheus.")
		return 0, 0, fmt.Errorf("no successful ingress data returned from Prometheus")
	}

	// Fetch egress usage
	egressResp, err := http.Get(egressURL)
	if err != nil {
		log.Error("Error querying Prometheus for egress: %v", err)
		return ingress, 0, err // Return ingress value if egress fails
	}
	defer egressResp.Body.Close()

	// Read the egress response body
	egressBody, err := ioutil.ReadAll(egressResp.Body)
	if err != nil {
		log.Error("Error reading egress response from Prometheus: %v", err)
		return ingress, 0, err
	}

	// Unmarshal the egress response into the PrometheusResponse struct
	var egressResponse PrometheusResponse
	if err := json.Unmarshal(egressBody, &egressResponse); err != nil {
		log.Error("Error unmarshalling egress Prometheus response: %v", err)
		return ingress, 0, err
	}

	// Extract egress value
	var egress float64
	if egressResponse.Status == "success" && len(egressResponse.Data.Result) > 0 {
		if egressValue, ok := egressResponse.Data.Result[0].Value[1].(string); ok {
			if _, err := fmt.Sscanf(egressValue, "%f", &egress); err != nil {
				log.Error("Error parsing egress network usage: %v", err)
				return ingress, 0, err
			}
		}
	} else {
		log.Error("No successful egress data returned from Prometheus.")
		return ingress, 0, fmt.Errorf("no successful egress data returned from Prometheus")
	}

	return ingress, egress, nil
}
