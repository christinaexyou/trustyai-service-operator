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
func (in *ChunkerSpec) DeepCopyInto(out *ChunkerSpec) {
	*out = *in
	out.Service = in.Service
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChunkerSpec.
func (in *ChunkerSpec) DeepCopy() *ChunkerSpec {
	if in == nil {
		return nil
	}
	out := new(ChunkerSpec)
	in.DeepCopyInto(out)
	return out
}

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
func (in *DetectorSpec) DeepCopyInto(out *DetectorSpec) {
	*out = *in
	out.Service = in.Service
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DetectorSpec.
func (in *DetectorSpec) DeepCopy() *DetectorSpec {
	if in == nil {
		return nil
	}
	out := new(DetectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GeneratorSpec) DeepCopyInto(out *GeneratorSpec) {
	*out = *in
	out.Service = in.Service
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GeneratorSpec.
func (in *GeneratorSpec) DeepCopy() *GeneratorSpec {
	if in == nil {
		return nil
	}
	out := new(GeneratorSpec)
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
	out.Generator = in.Generator
	if in.Chunkers != nil {
		in, out := &in.Chunkers, &out.Chunkers
		*out = make([]ChunkerSpec, len(*in))
		copy(*out, *in)
	}
	if in.Detectors != nil {
		in, out := &in.Detectors, &out.Detectors
		*out = make([]DetectorSpec, len(*in))
		copy(*out, *in)
	}
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
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	out.TLS = in.TLS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSSpec) DeepCopyInto(out *TLSSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSSpec.
func (in *TLSSpec) DeepCopy() *TLSSpec {
	if in == nil {
		return nil
	}
	out := new(TLSSpec)
	in.DeepCopyInto(out)
	return out
}
