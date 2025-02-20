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
