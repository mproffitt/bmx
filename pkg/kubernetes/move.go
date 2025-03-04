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
	"k8s.io/client-go/tools/clientcmd"
)

// Move context between sessions

func MoveContext(name, origfile, newfile string) error {
	if origfile == newfile {
		return nil
	}
	fullname := GetFullName(name, origfile)
	_, originalConfig, err := getApiConfig(origfile)
	if err != nil {
		return err
	}

	_, nc, err := getApiConfig(newfile)
	if err != nil {
		return err
	}

	newConfig := *nc

	context := *originalConfig.Contexts[fullname]
	authinfo := *originalConfig.AuthInfos[context.AuthInfo]
	cluster := *originalConfig.Clusters[context.Cluster]
	newConfig.Contexts[fullname] = &context
	newConfig.AuthInfos[context.AuthInfo] = &authinfo
	newConfig.Clusters[context.Cluster] = &cluster

	err = clientcmd.WriteToFile(newConfig, newfile)
	if err != nil {
		return err
	}
	return DeleteContext(name, origfile)
}
