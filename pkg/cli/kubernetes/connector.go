/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type connector struct {
}

//NewConnector returns new implementation of Connector
func NewConnector() *connector {
	return &connector{}
}

var clientSet *kubernetes.Clientset

func (c connector) Connect(clientGetter genericclioptions.RESTClientGetter) (string, error) {
	restConfig, err := clientGetter.ToRESTConfig()
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return "", err
	}

	clientSet = clientset
	ns, _, err := clientGetter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return "", err
	}

	return ns, nil

	// homeDirectory, err := homedir.Dir()
	// if err != nil {
	// 	return err
	// }
	// kubeconfig := filepath.Join(homeDirectory, ".kube", "config")
	// config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	// if err != nil {
	// 	return err
	// }

	// // create the clientset
	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	return err
	// }

	// clientSet = clientset
	// return nil
}
