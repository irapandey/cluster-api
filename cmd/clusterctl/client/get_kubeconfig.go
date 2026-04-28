/*
Copyright 2020 The Kubernetes Authors.

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

package client

import (
	"context"

	"github.com/pkg/errors"

	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/cluster"
)

// GetKubeconfigOptions carries all the options supported by GetKubeconfig.
type GetKubeconfigOptions struct {
	// Kubeconfig defines the kubeconfig to use for accessing the management cluster. If empty,
	// default rules for kubeconfig discovery will be used.
	Kubeconfig Kubeconfig

	// Namespace is the namespace in which secret is placed.
	Namespace string

	// WorkloadClusterName is the name of the workload cluster.
	WorkloadClusterName string
}

func (c *clusterctlClient) GetKubeconfig(ctx context.Context, options GetKubeconfigOptions) (string, error) {
	// gets access to the management cluster
	clusterClient, err := c.clusterClientFactory(ClusterClientFactoryInput{Kubeconfig: options.Kubeconfig})
	if err != nil {
		return "", err
	}

	// Ensure this command only runs against management clusters with the current Cluster API contract;
	// temporarily we also allow v1beta1 to allow transition to the current Cluster API contract.
	if err := clusterClient.ProviderInventory().CheckCAPIContract(ctx, cluster.AllowCAPIContract{Contract: clusterv1beta1.GroupVersion.Version}); err != nil {
		return "", err
	}

	if options.Namespace == "" {
		currentNamespace, err := clusterClient.Proxy().CurrentNamespace()
		if err != nil {
			return "", err
		}
		if currentNamespace == "" {
			return "", errors.New("failed to identify the current namespace. Please specify the namespace where the workload cluster exists")
		}
		options.Namespace = currentNamespace
	}

	return clusterClient.WorkloadCluster().GetKubeconfig(ctx, options.WorkloadClusterName, options.Namespace)
}
