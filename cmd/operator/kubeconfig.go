package main

import (
	"os"
	"path/filepath"

	"github.com/grafana/grafana-app-sdk/plugin/kubeconfig"
)

func LoadKubeconfig(val, path string) (kubeconfig.NamespacedConfig, error) {
	switch {
	case val != "":
		return LoadKubeConfigFromEnv(val)
	case path != "":
		return LoadKubeConfigFromFile(path)
	default:
		return LoadInClusterConfig()
	}
}

// LoadInClusterConfig loads a kubernetes in-cluster config.
// Since the in-cluster config doesn't have a namespace, it defaults to "default"
func LoadInClusterConfig() (kubeconfig.NamespacedConfig, error) {
	return LoadKubeConfigFromEnv("cluster")
}

// LoadKubeConfigFromFile loads a NamespacedConfig from a file on-disk (such as a mounted secret)
func LoadKubeConfigFromFile(path string) (kubeconfig.NamespacedConfig, error) {
	var res kubeconfig.NamespacedConfig

	path, err := filepath.Abs(path)
	if err != nil {
		return res, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return res, nil
	}

	return LoadKubeConfigFromEnv(string(bytes))
}

// LoadKubeConfigFromEnv loads a NamespacedConfig from the value of an environment variable
func LoadKubeConfigFromEnv(val string) (kubeconfig.NamespacedConfig, error) {
	var res kubeconfig.NamespacedConfig

	if err := kubeconfig.NewLoader().Load(val, "default", &res); err != nil {
		return res, err
	}

	return res, nil
}
