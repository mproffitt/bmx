package kubernetes

import (
	"encoding/json"
	"errors"
	"os/exec"

	bmx "github.com/mproffitt/bmx/pkg/exec"
)

type TeleportCluster struct {
	ClusterName string            `json:"kube_cluster_name"`
	Labels      map[string]string `json:"labels"`
	Selected    bool              `json:"selected"`
}

func TeleportClusterList() ([]string, error) {
	clusters := make([]string, 0)
	tsh, err := exec.LookPath("tsh")
	if err != nil {
		return clusters, errors.ErrUnsupported
	}

	out, _, err := bmx.Exec(tsh, []string{
		"kube", "ls", "-f", "json",
	})
	if err != nil {
		return clusters, err
	}

	var contents []TeleportCluster
	err = json.Unmarshal([]byte(out), &contents)
	if err != nil {
		return clusters, err
	}

	for _, v := range contents {
		clusters = append(clusters, v.ClusterName)
	}

	return clusters, nil
}

func TeleportClusterLogin(cluster string) error {
	tsh, err := exec.LookPath("tsh")
	if err != nil {
		return errors.ErrUnsupported
	}

	return bmx.ExecSilent(tsh, []string{
		"kube", "login", cluster,
	})
}
