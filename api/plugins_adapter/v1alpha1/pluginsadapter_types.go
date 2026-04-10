package v1alpha1

import (
	"github.com/trustyai-explainability/trustyai-service-operator/api/common"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PluginConfig holds information related to a single configuration of the Plugins Adapter server
type PluginConfig struct {
	// Name sets of the id of this particular config within the Plugins Adapter server. This will create a directory called /app/config/$Name. Since it $Name will be used a directory, it must only contain alphanumeric characters, dashes, and underscores.
	Name string `json:"name,omitempty"`
	//ConfigMaps is a list of configmaps that comprise the configuration. All files from these configmaps will be mounted within /app/config/$Name
	ConfigMaps []string `json:"configMaps,omitempty"`
	//The Default flag determines whether config is treated as the default config for the nemo-server. If no config is set to default, the first entry in NemoConfigs will be used as the default
	Default bool `json:"default,omitempty"`
}

// GatewayConfig holds information related to a single MCP Gateway CR
type GatewayConfig struct {
	// Name of the MCP Gateway CR
	Name string `json:"name,omitempty"`
	// Namespace of the MCP Gateway CR
	Namespace string `json:"namespace,omitempty"`
}

// NeMoGuardrailsConfig holds information related to a single NeMo Guardrails CR
type NeMoGuardrailsConfig struct {
	// Name of the NeMo Guardrails CR
	Name string `json:"name,omitempty"`
	// Namespace of the NeMo Guardrails CR
	Namespace string `json:"namespace,omitempty"`
}

// PluginsAdapterSpec defines the desired state of PluginsAdapter
type PluginsAdapterSpec struct {
	// PluginsConfigs should be the names of the configmaps containing Plugins Adapter server configuration files. All files in PluginsConfigs will be mounted to /app/config/$Name
	PluginsConfigs []PluginConfig `json:"pluginsConfigs"`

	// NemoConfig references an existing NeMo Guardrails CR whose Route will be
	// injected as DEFAULT_GUARDRAILS_SERVER_URL.
	NemoConfig *NeMoGuardrailsConfig `json:"nemoConfig,omitempty"`

	// GatewayConfig references the MCP Gateway whose namespace is used for
	// route / service validation.
	GatewayConfig *GatewayConfig `json:"gatewayConfig,omitempty"`

	// Env defines extra environment variables for the main container.
	// +optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

// PluginsAdapterStatus defines the observed state of PluginsAdapter
type PluginsAdapterStatus struct {
	Phase string `json:"phase,omitempty"`

	// Conditions describes the state of the PluginsAdapter resource.
	// +optional
	Conditions []common.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PluginsAdapter is the Schema for the pluginsadapters API
type PluginsAdapter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PluginsAdapterSpec   `json:"spec,omitempty"`
	Status PluginsAdapterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PluginsAdapterList contains a list of PluginsAdapter
type PluginsAdapterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PluginsAdapter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PluginsAdapter{}, &PluginsAdapterList{})
}
