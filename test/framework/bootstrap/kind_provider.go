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

package bootstrap

import (
	"context"
	"fmt"
	"os"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	kindv1 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kind "sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/exec"

	"sigs.k8s.io/cluster-api/test/framework/internal/log"
	"sigs.k8s.io/cluster-api/test/infrastructure/container"
)

const (
	// DefaultNodeImageRepository is the default node image repository to be used for testing.
	DefaultNodeImageRepository = "kindest/node"

	// DefaultNodeImageVersion is the default Kubernetes version to be used for creating a kind cluster.
	DefaultNodeImageVersion = "v1.34.0@sha256:7416a61b42b1662ca6ca89f02028ac133a309a2a30ba309614e8ec94d976dc5a"
)

// KindClusterOption is a NewKindClusterProvider option.
type KindClusterOption interface {
	apply(*KindClusterProvider)
}

type kindClusterOptionAdapter func(*KindClusterProvider)

func (adapter kindClusterOptionAdapter) apply(kindClusterProvider *KindClusterProvider) {
	adapter(kindClusterProvider)
}

// WithNodeImage implements a New Option that instruct the kindClusterProvider to use a specific node image / Kubernetes version.
func WithNodeImage(image string) KindClusterOption {
	return kindClusterOptionAdapter(func(k *KindClusterProvider) {
		k.nodeImage = image
	})
}

// WithDockerSockMount implements a New Option that instruct the kindClusterProvider to mount /var/run/docker.sock into
// the new kind cluster.
func WithDockerSockMount() KindClusterOption {
	return kindClusterOptionAdapter(func(k *KindClusterProvider) {
		k.withDockerSock = true
	})
}

// WithIPv6Family implements a New Option that instruct the kindClusterProvider to set the IPFamily to IPv6 in
// the new kind cluster.
func WithIPv6Family() KindClusterOption {
	return kindClusterOptionAdapter(func(k *KindClusterProvider) {
		k.ipFamily = kindv1.IPv6Family
	})
}

// WithDualStackFamily implements a New Option that instruct the kindClusterProvider to set the IPFamily to dual in
// the new kind cluster.
func WithDualStackFamily() KindClusterOption {
	return kindClusterOptionAdapter(func(k *KindClusterProvider) {
		k.ipFamily = kindv1.DualStackFamily
	})
}

// WithExtraPortMappings implements a New Option that instruct the kindClusterProvider to set extra port forward mappings.
func WithExtraPortMappings(mappings []kindv1.PortMapping) KindClusterOption {
	return kindClusterOptionAdapter(func(k *KindClusterProvider) {
		k.extraPortMappings = mappings
	})
}

// LogFolder implements a New Option that instruct the kindClusterProvider to dump bootstrap logs in a folder in case of errors.
func LogFolder(path string) KindClusterOption {
	return kindClusterOptionAdapter(func(k *KindClusterProvider) {
		k.logFolder = path
	})
}

// NewKindClusterProvider returns a ClusterProvider that can create a kind cluster.
func NewKindClusterProvider(name string, options ...KindClusterOption) *KindClusterProvider {
	Expect(name).ToNot(BeEmpty(), "name is required for NewKindClusterProvider")

	clusterProvider := &KindClusterProvider{
		name: name,
	}
	for _, option := range options {
		option.apply(clusterProvider)
	}
	return clusterProvider
}

// KindClusterProvider implements a ClusterProvider that can create a kind cluster.
type KindClusterProvider struct {
	name              string
	withDockerSock    bool
	kubeconfigPath    string
	nodeImage         string
	ipFamily          kindv1.ClusterIPFamily
	logFolder         string
	extraPortMappings []kindv1.PortMapping
}

// Create a Kubernetes cluster using kind.
func (k *KindClusterProvider) Create(ctx context.Context) {
	Expect(ctx).NotTo(BeNil(), "ctx is required for Create")

	// Sets the kubeconfig path to a temp file.
	// NB. the ClusterProvider is responsible for the cleanup of this file
	f, err := os.CreateTemp("", "e2e-kind")
	Expect(err).ToNot(HaveOccurred(), "Failed to create kubeconfig file for the kind cluster %q", k.name)
	k.kubeconfigPath = f.Name()

	// Creates the kind cluster
	k.createKindCluster()
}

// createKindCluster calls the kind library taking care of passing options for:
// - use a dedicated kubeconfig file (test should not alter the user environment)
// - if required, mount /var/run/docker.sock.
func (k *KindClusterProvider) createKindCluster() {
	kindCreateOptions := []kind.CreateOption{
		kind.CreateWithKubeconfigPath(k.kubeconfigPath),
	}

	cfg := &kindv1.Cluster{
		TypeMeta: kindv1.TypeMeta{
			APIVersion: "kind.x-k8s.io/v1alpha4",
			Kind:       "Cluster",
		},
		Nodes: []kindv1.Node{
			{
				Role:              kindv1.ControlPlaneRole,
				ExtraPortMappings: k.extraPortMappings,
			},
		},
	}

	if k.ipFamily == kindv1.IPv6Family {
		cfg.Networking.IPFamily = kindv1.IPv6Family
	}
	if k.ipFamily == kindv1.DualStackFamily {
		cfg.Networking.IPFamily = kindv1.DualStackFamily
	}
	kindv1.SetDefaultsCluster(cfg)

	if k.withDockerSock {
		cfg.Nodes[0].ExtraMounts = append(cfg.Nodes[0].ExtraMounts, kindv1.Mount{
			HostPath:      "/var/run/docker.sock",
			ContainerPath: "/var/run/docker.sock",
		})
	}

	kindCreateOptions = append(kindCreateOptions, kind.CreateWithV1Alpha4Config(cfg))

	nodeImage := fmt.Sprintf("%s:%s", DefaultNodeImageRepository, DefaultNodeImageVersion)
	if k.nodeImage != "" {
		nodeImage = k.nodeImage
	}
	kindCreateOptions = append(
		kindCreateOptions,
		kind.CreateWithNodeImage(nodeImage),
		kind.CreateWithRetain(true))

	provider := kind.NewProvider(kind.ProviderWithLogger(cmd.NewLogger()))
	err := provider.Create(k.name, kindCreateOptions...)
	if err != nil {
		// if requested, dump kind logs
		if k.logFolder != "" {
			if err := provider.CollectLogs(k.name, k.logFolder); err != nil {
				log.Logf("Failed to collect logs from kind: %v", err)
			}
		}

		errStr := fmt.Sprintf("Failed to create kind cluster %q: %v", k.name, err)
		// Extract the details of the RunError, if the cluster creation was triggered by a RunError.
		var runErr *exec.RunError
		if errors.As(err, &runErr) {
			errStr += "\n" + string(runErr.Output)
		}
		Expect(err).ToNot(HaveOccurred(), errStr)
	}
}

// GetKubeconfigPath returns the path to the kubeconfig file for the cluster.
func (k *KindClusterProvider) GetKubeconfigPath() string {
	return k.kubeconfigPath
}

// Dispose the kind cluster and its kubeconfig file.
// This method attempts to clean up the kind cluster and all associated containers.
// If the normal kind delete operation fails, it will attempt to manually clean up
// containers to prevent flaky tests due to leftover containers.
func (k *KindClusterProvider) Dispose(ctx context.Context) {
	Expect(ctx).NotTo(BeNil(), "ctx is required for Dispose")

	if err := kind.NewProvider().Delete(k.name, k.kubeconfigPath); err != nil {
		log.Logf("Deleting the kind cluster %q failed: %v. Attempting to clean up containers manually.", k.name, err)

		// If kind delete fails, try to clean up containers manually to prevent flaky tests
		// This addresses issue #12578 where containers were left running after test failures
		if err := k.forceCleanupContainers(ctx); err != nil {
			log.Logf("Failed to force cleanup containers for cluster %q: %v. You may need to remove this by hand.", k.name, err)
		}
	}
	if err := os.Remove(k.kubeconfigPath); err != nil {
		log.Logf("Deleting the kubeconfig file %q file. You may need to remove this by hand.", k.kubeconfigPath)
	}
}

// forceCleanupContainers attempts to manually clean up containers associated with this kind cluster
// when the normal kind delete operation fails. This helps prevent flaky tests due to leftover containers.
func (k *KindClusterProvider) forceCleanupContainers(ctx context.Context) error {
	// Get Docker runtime client
	containerRuntime, err := container.NewDockerClient()
	if err != nil {
		return errors.Wrap(err, "failed to get Docker runtime client")
	}
	ctx = container.RuntimeInto(ctx, containerRuntime)

	// Find all containers with the kind cluster name prefix
	filters := container.FilterBuilder{}
	filters.AddKeyValue("name", k.name)
	containers, err := containerRuntime.ListContainers(ctx, filters)
	if err != nil {
		return errors.Wrap(err, "failed to list containers")
	}

	// Also look for containers with the kind cluster name pattern (e.g., clusterctl-upgrade-management-2mbke1-control-plane)
	// This matches the pattern seen in the failing test logs
	kindPattern := k.name + "-"
	filtersPattern := container.FilterBuilder{}
	filtersPattern.AddKeyValue("name", kindPattern)
	patternContainers, err := containerRuntime.ListContainers(ctx, filtersPattern)
	if err != nil {
		log.Logf("Warning: failed to list containers with pattern %q: %v", kindPattern, err)
	} else {
		containers = append(containers, patternContainers...)
	}

	// Also look for any containers that might be related to this cluster by checking labels
	// Kind clusters often have labels that can help identify related containers
	labelFilters := container.FilterBuilder{}
	labelFilters.AddKeyValue("label", "io.x-k8s.kind.cluster="+k.name)
	labelContainers, err := containerRuntime.ListContainers(ctx, labelFilters)
	if err != nil {
		log.Logf("Warning: failed to list containers with cluster label %q: %v", k.name, err)
	} else {
		containers = append(containers, labelContainers...)
	}

	// Remove duplicates
	containerMap := make(map[string]container.Container)
	for _, c := range containers {
		containerMap[c.Name] = c
	}

	if len(containerMap) == 0 {
		log.Logf("No containers found for kind cluster %q", k.name)
		return nil
	}

	// Force delete all found containers
	var deleteErrors []error
	for _, c := range containerMap {
		log.Logf("Force deleting container: %s (Image: %s, Status: %s)", c.Name, c.Image, c.Status)
		if err := containerRuntime.DeleteContainer(ctx, c.Name); err != nil {
			deleteErrors = append(deleteErrors, errors.Wrapf(err, "failed to delete container %s", c.Name))
		}
	}

	if len(deleteErrors) > 0 {
		return errors.Wrap(kerrors.NewAggregate(deleteErrors), "failed to delete some containers")
	}

	log.Logf("Successfully cleaned up %d containers for kind cluster %q", len(containerMap), k.name)
	return nil
}
