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

package v1alpha4

import (
	"unsafe"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachineryconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta2"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
)

func (src *Cluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*clusterv1.Cluster)

	if err := Convert_v1alpha4_Cluster_To_v1beta2_Cluster(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1alpha4 conditions should not be automatically be converted into v1beta2 conditions.
	dst.Status.Conditions = nil

	// Move legacy conditions (v1alpha4), failureReason and failureMessage to the deprecated field.
	if src.Status.Conditions != nil || src.Status.FailureReason != nil || src.Status.FailureMessage != nil {
		dst.Status.Deprecated = &clusterv1.ClusterDeprecatedStatus{}
		dst.Status.Deprecated.V1Beta1 = &clusterv1.ClusterV1Beta1DeprecatedStatus{}
		if src.Status.Conditions != nil {
			Convert_v1alpha4_Conditions_To_v1beta2_Deprecated_V1Beta1_Conditions(&src.Status.Conditions, &dst.Status.Deprecated.V1Beta1.Conditions)
		}
		dst.Status.Deprecated.V1Beta1.FailureReason = src.Status.FailureReason
		dst.Status.Deprecated.V1Beta1.FailureMessage = src.Status.FailureMessage
	}

	// Move ControlPlaneReady and InfrastructureReady to Initialization
	if src.Status.ControlPlaneReady || src.Status.InfrastructureReady {
		if dst.Status.Initialization == nil {
			dst.Status.Initialization = &clusterv1.ClusterInitializationStatus{}
		}
		dst.Status.Initialization.ControlPlaneInitialized = src.Status.ControlPlaneReady
		dst.Status.Initialization.InfrastructureProvisioned = src.Status.InfrastructureReady
	}

	// Manually restore data.
	restored := &clusterv1.Cluster{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}

	dst.Spec.AvailabilityGates = restored.Spec.AvailabilityGates
	if restored.Spec.Topology != nil {
		if dst.Spec.Topology == nil {
			dst.Spec.Topology = &clusterv1.Topology{}
		}
		dst.Spec.Topology.ClassRef.Namespace = restored.Spec.Topology.ClassRef.Namespace
		dst.Spec.Topology.Variables = restored.Spec.Topology.Variables
		dst.Spec.Topology.ControlPlane.Variables = restored.Spec.Topology.ControlPlane.Variables

		if restored.Spec.Topology.ControlPlane.MachineHealthCheck != nil {
			dst.Spec.Topology.ControlPlane.MachineHealthCheck = restored.Spec.Topology.ControlPlane.MachineHealthCheck
		}

		if restored.Spec.Topology.ControlPlane.NodeDrainTimeout != nil {
			dst.Spec.Topology.ControlPlane.NodeDrainTimeout = restored.Spec.Topology.ControlPlane.NodeDrainTimeout
		}

		if restored.Spec.Topology.ControlPlane.NodeVolumeDetachTimeout != nil {
			dst.Spec.Topology.ControlPlane.NodeVolumeDetachTimeout = restored.Spec.Topology.ControlPlane.NodeVolumeDetachTimeout
		}

		if restored.Spec.Topology.ControlPlane.NodeDeletionTimeout != nil {
			dst.Spec.Topology.ControlPlane.NodeDeletionTimeout = restored.Spec.Topology.ControlPlane.NodeDeletionTimeout
		}
		dst.Spec.Topology.ControlPlane.ReadinessGates = restored.Spec.Topology.ControlPlane.ReadinessGates

		if restored.Spec.Topology.Workers != nil {
			if dst.Spec.Topology.Workers == nil {
				dst.Spec.Topology.Workers = &clusterv1.WorkersTopology{}
			}
			for i := range restored.Spec.Topology.Workers.MachineDeployments {
				dst.Spec.Topology.Workers.MachineDeployments[i].FailureDomain = restored.Spec.Topology.Workers.MachineDeployments[i].FailureDomain
				dst.Spec.Topology.Workers.MachineDeployments[i].Variables = restored.Spec.Topology.Workers.MachineDeployments[i].Variables
				dst.Spec.Topology.Workers.MachineDeployments[i].ReadinessGates = restored.Spec.Topology.Workers.MachineDeployments[i].ReadinessGates
				dst.Spec.Topology.Workers.MachineDeployments[i].NodeDrainTimeout = restored.Spec.Topology.Workers.MachineDeployments[i].NodeDrainTimeout
				dst.Spec.Topology.Workers.MachineDeployments[i].NodeVolumeDetachTimeout = restored.Spec.Topology.Workers.MachineDeployments[i].NodeVolumeDetachTimeout
				dst.Spec.Topology.Workers.MachineDeployments[i].NodeDeletionTimeout = restored.Spec.Topology.Workers.MachineDeployments[i].NodeDeletionTimeout
				dst.Spec.Topology.Workers.MachineDeployments[i].MinReadySeconds = restored.Spec.Topology.Workers.MachineDeployments[i].MinReadySeconds
				dst.Spec.Topology.Workers.MachineDeployments[i].Strategy = restored.Spec.Topology.Workers.MachineDeployments[i].Strategy
				dst.Spec.Topology.Workers.MachineDeployments[i].MachineHealthCheck = restored.Spec.Topology.Workers.MachineDeployments[i].MachineHealthCheck
			}

			dst.Spec.Topology.Workers.MachinePools = restored.Spec.Topology.Workers.MachinePools
		}
	}
	dst.Status.Conditions = restored.Status.Conditions
	dst.Status.ControlPlane = restored.Status.ControlPlane
	dst.Status.Workers = restored.Status.Workers

	return nil
}

func (dst *Cluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*clusterv1.Cluster)

	if err := Convert_v1beta2_Cluster_To_v1alpha4_Cluster(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1beta2 conditions should not be automatically be converted into legacy conditions (v1alpha4).
	dst.Status.Conditions = nil

	// Retrieve legacy conditions (v1alpha4), failureReason and failureMessage from the deprecated field.
	if src.Status.Deprecated != nil {
		if src.Status.Deprecated.V1Beta1 != nil {
			if src.Status.Deprecated.V1Beta1.Conditions != nil {
				Convert_v1beta2_Deprecated_V1Beta1_Conditions_To_v1alpha4_Conditions(&src.Status.Deprecated.V1Beta1.Conditions, &dst.Status.Conditions)
			}
			dst.Status.FailureReason = src.Status.Deprecated.V1Beta1.FailureReason
			dst.Status.FailureMessage = src.Status.Deprecated.V1Beta1.FailureMessage
		}
	}

	// Move initialization to old fields
	if src.Status.Initialization != nil {
		dst.Status.ControlPlaneReady = src.Status.Initialization.ControlPlaneInitialized
		dst.Status.InfrastructureReady = src.Status.Initialization.InfrastructureProvisioned
	}

	// Preserve Hub data on down-conversion except for metadata
	if err := utilconversion.MarshalData(src, dst); err != nil {
		return err
	}

	return nil
}

func (src *ClusterClass) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*clusterv1.ClusterClass)

	if err := Convert_v1alpha4_ClusterClass_To_v1beta2_ClusterClass(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1alpha4 conditions should not be automatically be converted into v1beta2 conditions.
	dst.Status.Conditions = nil

	// Manually restore data.
	restored := &clusterv1.ClusterClass{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}

	dst.Spec.Patches = restored.Spec.Patches
	dst.Spec.Variables = restored.Spec.Variables
	dst.Spec.AvailabilityGates = restored.Spec.AvailabilityGates
	dst.Spec.ControlPlane.MachineHealthCheck = restored.Spec.ControlPlane.MachineHealthCheck
	dst.Spec.ControlPlane.ReadinessGates = restored.Spec.ControlPlane.ReadinessGates
	dst.Spec.ControlPlane.NamingStrategy = restored.Spec.ControlPlane.NamingStrategy
	dst.Spec.Infrastructure.NamingStrategy = restored.Spec.Infrastructure.NamingStrategy
	dst.Spec.ControlPlane.NodeDrainTimeout = restored.Spec.ControlPlane.NodeDrainTimeout
	dst.Spec.ControlPlane.NodeVolumeDetachTimeout = restored.Spec.ControlPlane.NodeVolumeDetachTimeout
	dst.Spec.ControlPlane.NodeDeletionTimeout = restored.Spec.ControlPlane.NodeDeletionTimeout
	dst.Spec.Workers.MachinePools = restored.Spec.Workers.MachinePools

	for i := range restored.Spec.Workers.MachineDeployments {
		dst.Spec.Workers.MachineDeployments[i].MachineHealthCheck = restored.Spec.Workers.MachineDeployments[i].MachineHealthCheck
		dst.Spec.Workers.MachineDeployments[i].ReadinessGates = restored.Spec.Workers.MachineDeployments[i].ReadinessGates
		dst.Spec.Workers.MachineDeployments[i].FailureDomain = restored.Spec.Workers.MachineDeployments[i].FailureDomain
		dst.Spec.Workers.MachineDeployments[i].NamingStrategy = restored.Spec.Workers.MachineDeployments[i].NamingStrategy
		dst.Spec.Workers.MachineDeployments[i].NodeDrainTimeout = restored.Spec.Workers.MachineDeployments[i].NodeDrainTimeout
		dst.Spec.Workers.MachineDeployments[i].NodeVolumeDetachTimeout = restored.Spec.Workers.MachineDeployments[i].NodeVolumeDetachTimeout
		dst.Spec.Workers.MachineDeployments[i].NodeDeletionTimeout = restored.Spec.Workers.MachineDeployments[i].NodeDeletionTimeout
		dst.Spec.Workers.MachineDeployments[i].MinReadySeconds = restored.Spec.Workers.MachineDeployments[i].MinReadySeconds
		dst.Spec.Workers.MachineDeployments[i].Strategy = restored.Spec.Workers.MachineDeployments[i].Strategy
	}
	dst.Status = restored.Status

	return nil
}

func (dst *ClusterClass) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*clusterv1.ClusterClass)

	if err := Convert_v1beta2_ClusterClass_To_v1alpha4_ClusterClass(src, dst, nil); err != nil {
		return err
	}

	// Preserve Hub data on down-conversion except for metadata
	if err := utilconversion.MarshalData(src, dst); err != nil {
		return err
	}

	return nil
}

func (src *Machine) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*clusterv1.Machine)

	if err := Convert_v1alpha4_Machine_To_v1beta2_Machine(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1alpha4 conditions should not be automatically be converted into v1beta2 conditions.
	dst.Status.Conditions = nil

	// Move legacy conditions (v1alpha4), failureReason and failureMessage to the deprecated field.
	if src.Status.Conditions != nil || src.Status.FailureReason != nil || src.Status.FailureMessage != nil {
		dst.Status.Deprecated = &clusterv1.MachineDeprecatedStatus{}
		dst.Status.Deprecated.V1Beta1 = &clusterv1.MachineV1Beta1DeprecatedStatus{}
		if src.Status.Conditions != nil {
			Convert_v1alpha4_Conditions_To_v1beta2_Deprecated_V1Beta1_Conditions(&src.Status.Conditions, &dst.Status.Deprecated.V1Beta1.Conditions)
		}
		dst.Status.Deprecated.V1Beta1.FailureReason = src.Status.FailureReason
		dst.Status.Deprecated.V1Beta1.FailureMessage = src.Status.FailureMessage
	}

	// Move BootstrapReady and InfrastructureReady to Initialization
	if src.Status.BootstrapReady || src.Status.InfrastructureReady {
		if dst.Status.Initialization == nil {
			dst.Status.Initialization = &clusterv1.MachineInitializationStatus{}
		}
		dst.Status.Initialization.BootstrapDataSecretCreated = src.Status.BootstrapReady
		dst.Status.Initialization.InfrastructureProvisioned = src.Status.InfrastructureReady
	}

	// Manually restore data.
	restored := &clusterv1.Machine{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}

	dst.Spec.ReadinessGates = restored.Spec.ReadinessGates
	dst.Spec.NodeDeletionTimeout = restored.Spec.NodeDeletionTimeout
	dst.Status.CertificatesExpiryDate = restored.Status.CertificatesExpiryDate
	dst.Spec.NodeVolumeDetachTimeout = restored.Spec.NodeVolumeDetachTimeout
	dst.Status.Deletion = restored.Status.Deletion
	dst.Status.Conditions = restored.Status.Conditions

	return nil
}

func (dst *Machine) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*clusterv1.Machine)

	if err := Convert_v1beta2_Machine_To_v1alpha4_Machine(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1beta2 conditions should not be automatically be converted into legacy conditions (v1alpha4).
	dst.Status.Conditions = nil

	// Retrieve legacy conditions (v1alpha4), failureReason and failureMessage from the deprecated field.
	if src.Status.Deprecated != nil {
		if src.Status.Deprecated.V1Beta1 != nil {
			if src.Status.Deprecated.V1Beta1.Conditions != nil {
				Convert_v1beta2_Deprecated_V1Beta1_Conditions_To_v1alpha4_Conditions(&src.Status.Deprecated.V1Beta1.Conditions, &dst.Status.Conditions)
			}
			dst.Status.FailureReason = src.Status.Deprecated.V1Beta1.FailureReason
			dst.Status.FailureMessage = src.Status.Deprecated.V1Beta1.FailureMessage
		}
	}

	// Move initialization to old fields
	if src.Status.Initialization != nil {
		dst.Status.BootstrapReady = src.Status.Initialization.BootstrapDataSecretCreated
		dst.Status.InfrastructureReady = src.Status.Initialization.InfrastructureProvisioned
	}

	// Preserve Hub data on down-conversion except for metadata
	if err := utilconversion.MarshalData(src, dst); err != nil {
		return err
	}

	return nil
}

func (src *MachineSet) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*clusterv1.MachineSet)

	if err := Convert_v1alpha4_MachineSet_To_v1beta2_MachineSet(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1alpha4 conditions should not be automatically be converted into v1beta2 conditions.
	dst.Status.Conditions = nil

	// Move legacy conditions (v1alpha4), failureReason, failureMessage and replica counters to the deprecated field.
	dst.Status.Deprecated = &clusterv1.MachineSetDeprecatedStatus{}
	dst.Status.Deprecated.V1Beta1 = &clusterv1.MachineSetV1Beta1DeprecatedStatus{}
	if src.Status.Conditions != nil {
		Convert_v1alpha4_Conditions_To_v1beta2_Deprecated_V1Beta1_Conditions(&src.Status.Conditions, &dst.Status.Deprecated.V1Beta1.Conditions)
	}
	dst.Status.Deprecated.V1Beta1.FailureReason = src.Status.FailureReason
	dst.Status.Deprecated.V1Beta1.FailureMessage = src.Status.FailureMessage
	dst.Status.Deprecated.V1Beta1.ReadyReplicas = src.Status.ReadyReplicas
	dst.Status.Deprecated.V1Beta1.AvailableReplicas = src.Status.AvailableReplicas
	dst.Status.Deprecated.V1Beta1.FullyLabeledReplicas = src.Status.FullyLabeledReplicas

	// Manually restore data.
	restored := &clusterv1.MachineSet{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}

	dst.Spec.Template.Spec.ReadinessGates = restored.Spec.Template.Spec.ReadinessGates
	dst.Spec.Template.Spec.NodeDeletionTimeout = restored.Spec.Template.Spec.NodeDeletionTimeout
	dst.Spec.Template.Spec.NodeVolumeDetachTimeout = restored.Spec.Template.Spec.NodeVolumeDetachTimeout
	dst.Status.Conditions = restored.Status.Conditions
	dst.Status.AvailableReplicas = restored.Status.AvailableReplicas
	dst.Status.ReadyReplicas = restored.Status.ReadyReplicas
	dst.Status.UpToDateReplicas = restored.Status.UpToDateReplicas

	if restored.Spec.MachineNamingStrategy != nil {
		dst.Spec.MachineNamingStrategy = restored.Spec.MachineNamingStrategy
	}

	return nil
}

func (dst *MachineSet) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*clusterv1.MachineSet)

	if err := Convert_v1beta2_MachineSet_To_v1alpha4_MachineSet(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1beta2 conditions should not be automatically be converted into legacy conditions (v1alpha4).
	dst.Status.Conditions = nil

	// Reset replica counters from autogenerated conversions
	// NOTE: replica counters with a new semantic should not be automatically be converted into old replica counters.
	dst.Status.AvailableReplicas = 0
	dst.Status.ReadyReplicas = 0

	// Retrieve legacy conditions (v1alpha4), failureReason, failureMessage and replica counters from the deprecated field.
	if src.Status.Deprecated != nil {
		if src.Status.Deprecated.V1Beta1 != nil {
			if src.Status.Deprecated.V1Beta1.Conditions != nil {
				Convert_v1beta2_Deprecated_V1Beta1_Conditions_To_v1alpha4_Conditions(&src.Status.Deprecated.V1Beta1.Conditions, &dst.Status.Conditions)
			}
			dst.Status.FailureReason = src.Status.Deprecated.V1Beta1.FailureReason
			dst.Status.FailureMessage = src.Status.Deprecated.V1Beta1.FailureMessage
			dst.Status.ReadyReplicas = src.Status.Deprecated.V1Beta1.ReadyReplicas
			dst.Status.AvailableReplicas = src.Status.Deprecated.V1Beta1.AvailableReplicas
			dst.Status.FullyLabeledReplicas = src.Status.Deprecated.V1Beta1.FullyLabeledReplicas
		}
	}

	// Preserve Hub data on down-conversion except for metadata
	return utilconversion.MarshalData(src, dst)
}

func (src *MachineDeployment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*clusterv1.MachineDeployment)

	if err := Convert_v1alpha4_MachineDeployment_To_v1beta2_MachineDeployment(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1alpha4 conditions should not be automatically be converted into v1beta2 conditions.
	dst.Status.Conditions = nil

	// Move legacy conditions (v1alpha4) and replica counters to the deprecated field.
	dst.Status.Deprecated = &clusterv1.MachineDeploymentDeprecatedStatus{}
	dst.Status.Deprecated.V1Beta1 = &clusterv1.MachineDeploymentV1Beta1DeprecatedStatus{}
	if src.Status.Conditions != nil {
		Convert_v1alpha4_Conditions_To_v1beta2_Deprecated_V1Beta1_Conditions(&src.Status.Conditions, &dst.Status.Deprecated.V1Beta1.Conditions)
	}
	dst.Status.Deprecated.V1Beta1.ReadyReplicas = src.Status.ReadyReplicas
	dst.Status.Deprecated.V1Beta1.AvailableReplicas = src.Status.AvailableReplicas
	dst.Status.Deprecated.V1Beta1.UpdatedReplicas = src.Status.UpdatedReplicas
	dst.Status.Deprecated.V1Beta1.UnavailableReplicas = src.Status.UnavailableReplicas

	// Manually restore data.
	restored := &clusterv1.MachineDeployment{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}

	dst.Spec.Template.Spec.ReadinessGates = restored.Spec.Template.Spec.ReadinessGates
	dst.Spec.Template.Spec.NodeDeletionTimeout = restored.Spec.Template.Spec.NodeDeletionTimeout
	dst.Spec.Template.Spec.NodeVolumeDetachTimeout = restored.Spec.Template.Spec.NodeVolumeDetachTimeout
	dst.Spec.RolloutAfter = restored.Spec.RolloutAfter

	if restored.Spec.Strategy != nil {
		if dst.Spec.Strategy == nil {
			dst.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{}
		}
		dst.Spec.Strategy.Remediation = restored.Spec.Strategy.Remediation
	}

	if restored.Spec.MachineNamingStrategy != nil {
		dst.Spec.MachineNamingStrategy = restored.Spec.MachineNamingStrategy
	}
	dst.Status.Conditions = restored.Status.Conditions
	dst.Status.AvailableReplicas = restored.Status.AvailableReplicas
	dst.Status.ReadyReplicas = restored.Status.ReadyReplicas
	dst.Status.UpToDateReplicas = restored.Status.UpToDateReplicas

	return nil
}

func (dst *MachineDeployment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*clusterv1.MachineDeployment)

	if err := Convert_v1beta2_MachineDeployment_To_v1alpha4_MachineDeployment(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1beta2 conditions should not be automatically be converted into legacy conditions (v1alpha4).
	dst.Status.Conditions = nil

	// Reset replica counters from autogenerated conversions
	// NOTE: replica counters with a new semantic should not be automatically be converted into old replica counters.
	dst.Status.AvailableReplicas = 0
	dst.Status.ReadyReplicas = 0

	// Retrieve legacy conditions (v1alpha4), failureReason, failureMessage and replica counters from the deprecated field.
	if src.Status.Deprecated != nil {
		if src.Status.Deprecated.V1Beta1 != nil {
			if src.Status.Deprecated.V1Beta1.Conditions != nil {
				Convert_v1beta2_Deprecated_V1Beta1_Conditions_To_v1alpha4_Conditions(&src.Status.Deprecated.V1Beta1.Conditions, &dst.Status.Conditions)
			}
			dst.Status.ReadyReplicas = src.Status.Deprecated.V1Beta1.ReadyReplicas
			dst.Status.AvailableReplicas = src.Status.Deprecated.V1Beta1.AvailableReplicas
			dst.Status.UpdatedReplicas = src.Status.Deprecated.V1Beta1.UpdatedReplicas
			dst.Status.UnavailableReplicas = src.Status.Deprecated.V1Beta1.UnavailableReplicas
		}
	}

	// Preserve Hub data on down-conversion except for metadata
	return utilconversion.MarshalData(src, dst)
}

func (src *MachineHealthCheck) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*clusterv1.MachineHealthCheck)

	if err := Convert_v1alpha4_MachineHealthCheck_To_v1beta2_MachineHealthCheck(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1alpha4 conditions should not be automatically be converted into v1beta2 conditions.
	dst.Status.Conditions = nil

	// Move legacy conditions (v1alpha4) to the deprecated field.
	if src.Status.Conditions != nil {
		dst.Status.Deprecated = &clusterv1.MachineHealthCheckDeprecatedStatus{}
		dst.Status.Deprecated.V1Beta1 = &clusterv1.MachineHealthCheckV1Beta1DeprecatedStatus{}
		Convert_v1alpha4_Conditions_To_v1beta2_Deprecated_V1Beta1_Conditions(&src.Status.Conditions, &dst.Status.Deprecated.V1Beta1.Conditions)
	}

	// Manually restore data.
	restored := &clusterv1.MachineHealthCheck{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}
	dst.Status.Conditions = restored.Status.Conditions

	return nil
}

func (dst *MachineHealthCheck) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*clusterv1.MachineHealthCheck)

	if err := Convert_v1beta2_MachineHealthCheck_To_v1alpha4_MachineHealthCheck(src, dst, nil); err != nil {
		return err
	}

	// Reset conditions from autogenerated conversions
	// NOTE: v1beta2 conditions should not be automatically be converted into legacy conditions (v1alpha4).
	dst.Status.Conditions = nil

	// Retrieve legacy conditions (v1alpha4) from the deprecated field.
	if src.Status.Deprecated != nil {
		if src.Status.Deprecated.V1Beta1 != nil {
			if src.Status.Deprecated.V1Beta1.Conditions != nil {
				Convert_v1beta2_Deprecated_V1Beta1_Conditions_To_v1alpha4_Conditions(&src.Status.Deprecated.V1Beta1.Conditions, &dst.Status.Conditions)
			}
		}
	}

	// Preserve Hub data on down-conversion except for metadata
	return utilconversion.MarshalData(src, dst)
}

func Convert_v1alpha4_MachineStatus_To_v1beta2_MachineStatus(in *MachineStatus, out *clusterv1.MachineStatus, s apimachineryconversion.Scope) error {
	// Status.version has been removed in v1beta1, thus requiring custom conversion function. the information will be dropped.
	return autoConvert_v1alpha4_MachineStatus_To_v1beta2_MachineStatus(in, out, s)
}

func Convert_v1beta2_ClusterClassSpec_To_v1alpha4_ClusterClassSpec(in *clusterv1.ClusterClassSpec, out *ClusterClassSpec, s apimachineryconversion.Scope) error {
	// spec.{variables,patches} has been added with v1beta1.
	return autoConvert_v1beta2_ClusterClassSpec_To_v1alpha4_ClusterClassSpec(in, out, s)
}

func Convert_v1beta2_InfrastructureClass_To_v1alpha4_LocalObjectTemplate(in *clusterv1.InfrastructureClass, out *LocalObjectTemplate, s apimachineryconversion.Scope) error {
	if in == nil {
		return nil
	}

	return autoConvert_v1beta2_LocalObjectTemplate_To_v1alpha4_LocalObjectTemplate(&in.LocalObjectTemplate, out, s)
}

func Convert_v1alpha4_LocalObjectTemplate_To_v1beta2_InfrastructureClass(in *LocalObjectTemplate, out *clusterv1.InfrastructureClass, s apimachineryconversion.Scope) error {
	if in == nil {
		return nil
	}

	return autoConvert_v1alpha4_LocalObjectTemplate_To_v1beta2_LocalObjectTemplate(in, &out.LocalObjectTemplate, s)
}

func Convert_v1beta2_MachineSpec_To_v1alpha4_MachineSpec(in *clusterv1.MachineSpec, out *MachineSpec, s apimachineryconversion.Scope) error {
	// spec.nodeDeletionTimeout was added in v1beta1.
	// ReadinessGates was added in v1beta1.
	return autoConvert_v1beta2_MachineSpec_To_v1alpha4_MachineSpec(in, out, s)
}

func Convert_v1beta2_MachineDeploymentSpec_To_v1alpha4_MachineDeploymentSpec(in *clusterv1.MachineDeploymentSpec, out *MachineDeploymentSpec, s apimachineryconversion.Scope) error {
	return autoConvert_v1beta2_MachineDeploymentSpec_To_v1alpha4_MachineDeploymentSpec(in, out, s)
}

func Convert_v1beta2_ClusterSpec_To_v1alpha4_ClusterSpec(in *clusterv1.ClusterSpec, out *ClusterSpec, s apimachineryconversion.Scope) error {
	// AvailabilityGates was added in v1beta1.
	return autoConvert_v1beta2_ClusterSpec_To_v1alpha4_ClusterSpec(in, out, s)
}

func Convert_v1beta2_ClusterStatus_To_v1alpha4_ClusterStatus(in *clusterv1.ClusterStatus, out *ClusterStatus, s apimachineryconversion.Scope) error {
	// V1Beta2 was added in v1beta1.
	return autoConvert_v1beta2_ClusterStatus_To_v1alpha4_ClusterStatus(in, out, s)
}

func Convert_v1beta2_Topology_To_v1alpha4_Topology(in *clusterv1.Topology, out *Topology, s apimachineryconversion.Scope) error {
	// spec.topology.variables has been added with v1beta1.
	if err := autoConvert_v1beta2_Topology_To_v1alpha4_Topology(in, out, s); err != nil {
		return err
	}

	out.Class = in.ClassRef.Name
	return nil
}

// Convert_v1beta2_MachineDeploymentTopology_To_v1alpha4_MachineDeploymentTopology is an autogenerated conversion function.
func Convert_v1beta2_MachineDeploymentTopology_To_v1alpha4_MachineDeploymentTopology(in *clusterv1.MachineDeploymentTopology, out *MachineDeploymentTopology, s apimachineryconversion.Scope) error {
	// MachineDeploymentTopology.FailureDomain has been added with v1beta1.
	return autoConvert_v1beta2_MachineDeploymentTopology_To_v1alpha4_MachineDeploymentTopology(in, out, s)
}

func Convert_v1beta2_MachineDeploymentClass_To_v1alpha4_MachineDeploymentClass(in *clusterv1.MachineDeploymentClass, out *MachineDeploymentClass, s apimachineryconversion.Scope) error {
	// machineDeploymentClass.machineHealthCheck has been added with v1beta1.
	return autoConvert_v1beta2_MachineDeploymentClass_To_v1alpha4_MachineDeploymentClass(in, out, s)
}

func Convert_v1beta2_ControlPlaneClass_To_v1alpha4_ControlPlaneClass(in *clusterv1.ControlPlaneClass, out *ControlPlaneClass, s apimachineryconversion.Scope) error {
	// controlPlaneClass.machineHealthCheck has been added with v1beta1.
	return autoConvert_v1beta2_ControlPlaneClass_To_v1alpha4_ControlPlaneClass(in, out, s)
}

func Convert_v1alpha4_Topology_To_v1beta2_Topology(in *Topology, out *clusterv1.Topology, s apimachineryconversion.Scope) error {
	if err := autoConvert_v1alpha4_Topology_To_v1beta2_Topology(in, out, s); err != nil {
		return err
	}

	out.ClassRef.Name = in.Class
	return nil
}

func Convert_v1beta2_ControlPlaneTopology_To_v1alpha4_ControlPlaneTopology(in *clusterv1.ControlPlaneTopology, out *ControlPlaneTopology, s apimachineryconversion.Scope) error {
	// controlPlaneTopology.nodeDrainTimeout has been added with v1beta1.
	return autoConvert_v1beta2_ControlPlaneTopology_To_v1alpha4_ControlPlaneTopology(in, out, s)
}

func Convert_v1beta2_MachineStatus_To_v1alpha4_MachineStatus(in *clusterv1.MachineStatus, out *MachineStatus, s apimachineryconversion.Scope) error {
	// MachineStatus.CertificatesExpiryDate has been added in v1beta1.
	// V1Beta2 was added in v1beta1.
	return autoConvert_v1beta2_MachineStatus_To_v1alpha4_MachineStatus(in, out, s)
}

func Convert_v1beta2_ClusterClass_To_v1alpha4_ClusterClass(in *clusterv1.ClusterClass, out *ClusterClass, s apimachineryconversion.Scope) error {
	// ClusterClass.Status has been added in v1beta1.
	return autoConvert_v1beta2_ClusterClass_To_v1alpha4_ClusterClass(in, out, s)
}

func Convert_v1beta2_WorkersClass_To_v1alpha4_WorkersClass(in *clusterv1.WorkersClass, out *WorkersClass, s apimachineryconversion.Scope) error {
	// WorkersClass.MachinePools has been added in v1beta1.
	return autoConvert_v1beta2_WorkersClass_To_v1alpha4_WorkersClass(in, out, s)
}

func Convert_v1beta2_WorkersTopology_To_v1alpha4_WorkersTopology(in *clusterv1.WorkersTopology, out *WorkersTopology, s apimachineryconversion.Scope) error {
	// WorkersTopology.MachinePools has been added in v1beta1.
	return autoConvert_v1beta2_WorkersTopology_To_v1alpha4_WorkersTopology(in, out, s)
}

func Convert_v1beta2_MachineDeploymentStrategy_To_v1alpha4_MachineDeploymentStrategy(in *clusterv1.MachineDeploymentStrategy, out *MachineDeploymentStrategy, s apimachineryconversion.Scope) error {
	return autoConvert_v1beta2_MachineDeploymentStrategy_To_v1alpha4_MachineDeploymentStrategy(in, out, s)
}

func Convert_v1beta2_MachineSetSpec_To_v1alpha4_MachineSetSpec(in *clusterv1.MachineSetSpec, out *MachineSetSpec, s apimachineryconversion.Scope) error {
	return autoConvert_v1beta2_MachineSetSpec_To_v1alpha4_MachineSetSpec(in, out, s)
}

func Convert_v1beta2_MachineDeploymentStatus_To_v1alpha4_MachineDeploymentStatus(in *clusterv1.MachineDeploymentStatus, out *MachineDeploymentStatus, s apimachineryconversion.Scope) error {
	// V1Beta2 was added in v1beta1.
	return autoConvert_v1beta2_MachineDeploymentStatus_To_v1alpha4_MachineDeploymentStatus(in, out, s)
}

func Convert_v1beta2_MachineSetStatus_To_v1alpha4_MachineSetStatus(in *clusterv1.MachineSetStatus, out *MachineSetStatus, s apimachineryconversion.Scope) error {
	// V1Beta2 was added in v1beta1.
	return autoConvert_v1beta2_MachineSetStatus_To_v1alpha4_MachineSetStatus(in, out, s)
}

func Convert_v1beta2_MachineHealthCheckStatus_To_v1alpha4_MachineHealthCheckStatus(in *clusterv1.MachineHealthCheckStatus, out *MachineHealthCheckStatus, s apimachineryconversion.Scope) error {
	// V1Beta2 was added in v1beta1.
	return autoConvert_v1beta2_MachineHealthCheckStatus_To_v1alpha4_MachineHealthCheckStatus(in, out, s)
}

func Convert_v1alpha4_ClusterStatus_To_v1beta2_ClusterStatus(in *ClusterStatus, out *clusterv1.ClusterStatus, s apimachineryconversion.Scope) error {
	return autoConvert_v1alpha4_ClusterStatus_To_v1beta2_ClusterStatus(in, out, s)
}

func Convert_v1alpha4_MachineDeploymentStatus_To_v1beta2_MachineDeploymentStatus(in *MachineDeploymentStatus, out *clusterv1.MachineDeploymentStatus, s apimachineryconversion.Scope) error {
	return autoConvert_v1alpha4_MachineDeploymentStatus_To_v1beta2_MachineDeploymentStatus(in, out, s)
}

func Convert_v1alpha4_MachineSetStatus_To_v1beta2_MachineSetStatus(in *MachineSetStatus, out *clusterv1.MachineSetStatus, s apimachineryconversion.Scope) error {
	return autoConvert_v1alpha4_MachineSetStatus_To_v1beta2_MachineSetStatus(in, out, s)
}

func Convert_v1_Condition_To_v1alpha4_Condition(_ *metav1.Condition, _ *Condition, _ apimachineryconversion.Scope) error {
	// NOTE: v1beta2 conditions should not be automatically converted into legacy (v1alpha4) conditions.
	return nil
}

func Convert_v1alpha4_Condition_To_v1_Condition(_ *Condition, _ *metav1.Condition, _ apimachineryconversion.Scope) error {
	// NOTE: legacy (v1alpha4) conditions should not be automatically converted into v1beta2 conditions.
	return nil
}

func Convert_v1beta2_Deprecated_V1Beta1_Conditions_To_v1alpha4_Conditions(in *clusterv1.Conditions, out *Conditions) {
	*out = make(Conditions, len(*in))
	for i := range *in {
		(*out)[i] = *(*Condition)(unsafe.Pointer(&(*in)[i]))
	}
}

func Convert_v1alpha4_Conditions_To_v1beta2_Deprecated_V1Beta1_Conditions(in *Conditions, out *clusterv1.Conditions) {
	*out = make(clusterv1.Conditions, len(*in))
	for i := range *in {
		(*out)[i] = *(*clusterv1.Condition)(unsafe.Pointer(&(*in)[i]))
	}
}
