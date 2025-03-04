// Copyright (c) 2025 Martin Proffitt <mprooffitt@choclab.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
