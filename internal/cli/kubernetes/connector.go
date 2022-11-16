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
	"k8s.io/client-go/rest"
)

type connector struct {
}

// NewConnector returns new implementation of Connector
func NewConnector() *connector {
	return &connector{}
}

type KubeContext struct {
	ClientSet  kubernetes.Interface
	RestConfig *rest.Config
}

type ConnectionContext struct {
	KubeContext
	Namespace string
}

func (c connector) Connect(clientGetter genericclioptions.RESTClientGetter) (ConnectionContext, error) {
	restConfig, err := clientGetter.ToRESTConfig()
	if err != nil {
		return ConnectionContext{}, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return ConnectionContext{}, err
	}

	ns, _, err := clientGetter.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return ConnectionContext{}, err
	}

	return ConnectionContext{
		KubeContext: KubeContext{
			ClientSet:  clientset,
			RestConfig: restConfig,
		},
		Namespace: ns,
	}, nil
}
