//go:build !ignore_autogenerated

/*
Copyright 2023.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GuardrailsOrchestrator) DeepCopyInto(out *GuardrailsOrchestrator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GuardrailsOrchestrator.
func (in *GuardrailsOrchestrator) DeepCopy() *GuardrailsOrchestrator {
	if in == nil {
		return nil
	}
	out := new(GuardrailsOrchestrator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GuardrailsOrchestrator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GuardrailsOrchestratorList) DeepCopyInto(out *GuardrailsOrchestratorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GuardrailsOrchestrator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GuardrailsOrchestratorList.
func (in *GuardrailsOrchestratorList) DeepCopy() *GuardrailsOrchestratorList {
	if in == nil {
		return nil
	}
	out := new(GuardrailsOrchestratorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GuardrailsOrchestratorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GuardrailsOrchestratorSpec) DeepCopyInto(out *GuardrailsOrchestratorSpec) {
	*out = *in
	if in.OrchestratorConfig != nil {
		in, out := &in.OrchestratorConfig, &out.OrchestratorConfig
		*out = new(string)
		**out = **in
	}
	if in.SidecarGatewayConfig != nil {
		in, out := &in.SidecarGatewayConfig, &out.SidecarGatewayConfig
		*out = new(string)
		**out = **in
	}
	out.OtelExporter = in.OtelExporter
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GuardrailsOrchestratorSpec.
func (in *GuardrailsOrchestratorSpec) DeepCopy() *GuardrailsOrchestratorSpec {
	if in == nil {
		return nil
	}
	out := new(GuardrailsOrchestratorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GuardrailsOrchestratorStatus) DeepCopyInto(out *GuardrailsOrchestratorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GuardrailsOrchestratorStatus.
func (in *GuardrailsOrchestratorStatus) DeepCopy() *GuardrailsOrchestratorStatus {
	if in == nil {
		return nil
	}
	out := new(GuardrailsOrchestratorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OtelExporter) DeepCopyInto(out *OtelExporter) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OtelExporter.
func (in *OtelExporter) DeepCopy() *OtelExporter {
	if in == nil {
		return nil
	}
	out := new(OtelExporter)
	in.DeepCopyInto(out)
	return out
}
