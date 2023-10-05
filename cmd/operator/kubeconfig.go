package main

import (
	"fmt"

	"github.com/grafana/grafana-app-sdk/plugin/kubeconfig"
)

// LoadInClusterConfig loads a kubernetes in-cluster config.
// Since the in-cluster config doesn't have a namespace, it defaults to "default"
func LoadInClusterConfig() (kubeconfig.NamespacedConfig, error) {
	var res kubeconfig.NamespacedConfig

	if err := kubeconfig.NewLoader().Load("cluster", "default", &res); err != nil {
		return res, err
	}

	return res, nil
}

// LoadKubeConfigFromEnv loads a NamespacedConfig from the value of an environment variable
func LoadKubeConfigFromEnv() (*kubeconfig.NamespacedConfig, error) {
	// TODO
	return nil, fmt.Errorf("not implemented")
}

// LoadKubeConfigFromFile loads a NamespacedConfig from a file on-disk (such as a mounted secret)
func LoadKubeConfigFromFile() (*kubeconfig.NamespacedConfig, error) {
	// TODO
	return nil, fmt.Errorf("not implemented")
}
