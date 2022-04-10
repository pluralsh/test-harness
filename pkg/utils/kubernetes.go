package utils

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

const tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

func KubeConfig() (*rest.Config, error) {
	if exists(tokenFile) {
		return rest.InClusterConfig()
	}

	homedir, _ := os.UserHomeDir()
	conf := filepath.Join(homedir, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", conf)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
