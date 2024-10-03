package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

// PrometheusConfig holds the configuration for Prometheus.
type PrometheusConfig struct {
	URL string `yaml:"url"`
}

// Thresholds holds the configuration for usage thresholds.
type Thresholds struct {
	DiskUsage struct {
		Resize int `yaml:"resize"`
	} `yaml:"diskUsage"`
	NetworkUsage struct {
		Ingress struct {
			Scale int `yaml:"scale"`
		} `yaml:"ingress"`
	} `yaml:"networkUsage"`
}

// AutoscalerConfig holds the autoscaler settings and related configurations.
type AutoscalerConfig struct {
	DesiredReplicaCount int            `yaml:"desiredReplicaCount"`
	Interval            int            `yaml:"interval"`
	Prometheus          PrometheusConfig `yaml:"prometheus"`
	Thresholds          Thresholds      `yaml:"thresholds"`
}

// LoadConfig reads the configuration from the specified YAML file.
func LoadConfig(filePath string) (*AutoscalerConfig, error) {
	var config AutoscalerConfig
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
