package plugins_adapter

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	nemoguardrailsv1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/nemo_guardrails/v1alpha1"
	pluginsadapterv1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/plugins_adapter/v1alpha1"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/constants"
	templateParser "github.com/trustyai-explainability/trustyai-service-operator/controllers/plugins_adapter/templates"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/utils"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ContainerImages struct {
	PluginsAdapterImage string
}

type DeploymentConfig struct {
	PluginsAdapter  *pluginsadapterv1alpha1.PluginsAdapter
	ContainerImages ContainerImages
}

type loadedPluginConfigMap struct {
	refName string
	cm      *corev1.ConfigMap
}

const deploymentTemplateFilename = "deployment.tmpl.yaml"

// extractMCPRouteFromPluginYAML returns the first non-empty plugins[].mcp.url
// found in the YAML data (matching the shape of configmap.yaml).
func extractMCPRouteFromPluginYAML(data string) (string, bool) {
	var root struct {
		Plugins []struct {
			MCP struct {
				URL string `yaml:"url"`
			} `yaml:"mcp"`
		} `yaml:"plugins"`
	}
	if err := yaml.Unmarshal([]byte(strings.TrimSpace(data)), &root); err != nil {
		return "", false
	}
	for _, p := range root.Plugins {
		if u := strings.TrimSpace(p.MCP.URL); u != "" {
			return u, true
		}
	}
	return "", false
}

func (r *PluginsAdapterReconciler) mountPluginConfigs(ctx context.Context, pluginsAdapter *pluginsadapterv1alpha1.PluginsAdapter, deployment *appsv1.Deployment) error {
	if len(pluginsAdapter.Spec.PluginsConfigs) == 0 {
		return fmt.Errorf("no PluginsConfigs provided")
	}

	// Check if the MCP Gateway namespace is provided, if not use the plugins adapter namespace
	lookupNamespace := pluginsAdapter.Namespace
	if pluginsAdapter.Spec.GatewayConfig != nil && pluginsAdapter.Spec.GatewayConfig.Namespace != "" {
		lookupNamespace = pluginsAdapter.Spec.GatewayConfig.Namespace
	}

	// Determine the default config name
	defaultConfig := pluginsAdapter.Spec.PluginsConfigs[0].Name
	defaultAlreadyChosen := false
	for _, pluginConfig := range pluginsAdapter.Spec.PluginsConfigs {
		if pluginConfig.Default {
			if defaultAlreadyChosen {
				log.FromContext(ctx).Info(fmt.Sprintf(
					"warning: Two or more PluginConfigs have set default=true. Only '%s' will be used as default, as it was the first in the PluginConfig list to specify default=true.", defaultConfig))
			} else {
				defaultConfig = pluginConfig.Name
				defaultAlreadyChosen = true
			}
		}
	}
	if !defaultAlreadyChosen {
		log.FromContext(ctx).Info(fmt.Sprintf("no PluginConfigs were marked as default, using '%s' as default", defaultConfig))
	}

	// Load and mount all configs; validate the MCP URL only for the default config
	for _, pluginConfig := range pluginsAdapter.Spec.PluginsConfigs {
		if len(pluginConfig.ConfigMaps) == 0 {
			return fmt.Errorf("no configmaps provided inside PluginConfig=%s", pluginConfig.Name)
		}

		var combinedPluginConfigYAML strings.Builder
		loadedCMs := make([]loadedPluginConfigMap, 0, len(pluginConfig.ConfigMaps))
		for _, configCM := range pluginConfig.ConfigMaps {
			configmap := &corev1.ConfigMap{}
			if err := r.Client.Get(ctx, types.NamespacedName{Name: configCM, Namespace: pluginsAdapter.Namespace}, configmap); err != nil {
				utils.LogErrorRetrieving(ctx, err, "configmap", configCM, pluginsAdapter.Namespace)
				return err
			}
			for _, v := range configmap.Data {
				combinedPluginConfigYAML.WriteString(v)
				combinedPluginConfigYAML.WriteByte('\n')
			}
			loadedCMs = append(loadedCMs, loadedPluginConfigMap{refName: configCM, cm: configmap})
		}

		if pluginConfig.Name == defaultConfig {
			if rawURL, ok := extractMCPRouteFromPluginYAML(combinedPluginConfigYAML.String()); ok {
				if err := utils.ValidateMCPURLUsesAdmittedRoute(ctx, r.Client, rawURL, lookupNamespace); err != nil {
					return fmt.Errorf("default plugin config %q: %w", pluginConfig.Name, err)
				}
			}
		}

		for cmIdx, loaded := range loadedCMs {
			volumeName := fmt.Sprintf("%s-%s-volume", pluginConfig.Name, loaded.refName)
			utils.MountConfigMapToDeployment(loaded.cm, volumeName, deployment)
			mountPath := fmt.Sprintf("/app/config/%s/%d", pluginConfig.Name, cmIdx)
			volumeMount := corev1.VolumeMount{
				Name:      volumeName,
				MountPath: mountPath,
			}
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(
				deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
				volumeMount,
			)
		}
	}

	deployment.Spec.Template.Spec.Containers[0].Env = append(
		deployment.Spec.Template.Spec.Containers[0].Env,
		corev1.EnvVar{
			Name:  "CONFIG_ID",
			Value: defaultConfig,
		},
	)

	return nil
}

func (r *PluginsAdapterReconciler) createDeployment(ctx context.Context, pluginsAdapter *pluginsadapterv1alpha1.PluginsAdapter) (*appsv1.Deployment, error) {
	var containerImages ContainerImages

	// ===== get plugins adapter image from trustyai configmap ============================================================
	pluginsAdapterImage, err := utils.GetImageFromConfigMap(ctx, r.Client, pluginsAdapterImageKey, constants.ConfigMap, r.Namespace)
	if err != nil {
		utils.LogErrorRetrieving(ctx, err, "plugins adapter image from configmap", constants.ConfigMap, r.Namespace)
		return nil, err
	}
	if pluginsAdapterImage == "" {
		err = fmt.Errorf("configmap %s in namespace %s has empty value for key %s", constants.ConfigMap, r.Namespace, pluginsAdapterImageKey)
		utils.LogErrorRetrieving(ctx, err, "plugins adapter image from configmap", constants.ConfigMap, r.Namespace)
		return nil, err
	}
	containerImages.PluginsAdapterImage = pluginsAdapterImage
	log.FromContext(ctx).Info("using PluginsAdapterImage", "image", pluginsAdapterImage, "configmap", r.Namespace+":"+constants.ConfigMap)

	// ===== verify that the NeMo Guardrails server exists and its route is admitted ===========================
	var nemoRouteHost string
	if pluginsAdapter.Spec.NemoConfig != nil {
		nemoguardrails := &nemoguardrailsv1alpha1.NemoGuardrails{}
		nemoName := pluginsAdapter.Spec.NemoConfig.Name
		nemoNamespace := pluginsAdapter.Spec.NemoConfig.Namespace
		err = r.Client.Get(ctx, types.NamespacedName{Name: nemoName, Namespace: nemoNamespace}, nemoguardrails)
		if err != nil {
			return nil, fmt.Errorf("could not find NemoGuardrails %s/%s: %w", nemoNamespace, nemoName, err)
		}

		ready, err := utils.CheckRouteReady(ctx, r.Client, nemoguardrails.Name, nemoguardrails.Namespace)
		if err != nil {
			return nil, fmt.Errorf("could not verify route for NemoGuardrails %s/%s: %w", nemoNamespace, nemoName, err)
		}
		if !ready {
			return nil, fmt.Errorf("route for NemoGuardrails %s/%s is not admitted", nemoNamespace, nemoName)
		}

		route := &routev1.Route{}
		err = r.Client.Get(ctx, types.NamespacedName{Name: nemoguardrails.Name, Namespace: nemoguardrails.Namespace}, route)
		if err != nil {
			return nil, fmt.Errorf("could not get route for NemoGuardrails %s/%s: %w", nemoNamespace, nemoName, err)
		}
		nemoRouteHost = route.Spec.Host
	}

	// ===== create deployment definition ================================================================================
	deploymentConfig := DeploymentConfig{
		PluginsAdapter:  pluginsAdapter,
		ContainerImages: containerImages,
	}

	deployment, err := templateParser.ParseResource[*appsv1.Deployment](deploymentTemplateFilename, deploymentConfig, reflect.TypeOf(&appsv1.Deployment{}))
	if err != nil {
		utils.LogErrorParsing(ctx, err, "deployment template", pluginsAdapter.Name, pluginsAdapter.Namespace)
		return nil, err
	}
	if err := controllerutil.SetControllerReference(pluginsAdapter, deployment, r.Scheme); err != nil {
		utils.LogErrorControllerReference(ctx, err, "deployment", deployment.Name, deployment.Namespace)
		return nil, err
	}

	if nemoRouteHost != "" {
		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
			Name:  "DEFAULT_GUARDRAILS_SERVER_URL",
			Value: nemoRouteHost,
		})
	}

	// Add user plugin configs to deployment
	err = r.mountPluginConfigs(ctx, pluginsAdapter, deployment)
	if err != nil {
		return nil, err
	}

	// Add user environment variables
	if len(pluginsAdapter.Spec.Env) > 0 {
		log.FromContext(ctx).Info("Updating PluginsAdapter env with user-provided environment variables")
		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, pluginsAdapter.Spec.Env...)
	}

	return deployment, nil
}
