package nemo

import (
	"context"
	nemov1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/nemo/v1alpha1"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/constants"
	templateParser "github.com/trustyai-explainability/trustyai-service-operator/controllers/gorch/templates"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const deploymentTemplatePath = "deployment.tmpl.yaml"

type ContainerImages struct {
	NemoGuardrailsImage string
}

type DeploymentConfig struct {
	NemoGuardrail   *nemov1alpha1.NemoGuardrails
	ContainerImages ContainerImages
}

func (r *NemoGuardrailsReconciler) createDeployment(ctx context.Context, nemoGuardrails *nemov1alpha1.NemoGuardrails) *appsv1.Deployment {
	var containerImages ContainerImages

	nemoGuardrailsImage, err := utils.GetImageFromConfigMap(ctx, r.Client, nemoImageKey, constants.ConfigMap, r.Namespace)
	if nemoGuardrailsImage == "" || err != nil {
		log.FromContext(ctx).Error(err, "Error getting container image from ConfigMap.")
	}
	containerImages.NemoGuardrailsImage = nemoGuardrailsImage
	log.FromContext(ctx).Info("using NemoGuardrailsImage " + nemoGuardrailsImage + " " + "from config map " + r.Namespace + ":" + constants.ConfigMap)

	deploymentConfig := DeploymentConfig{
		NemoGuardrail:   nemoGuardrails,
		ContainerImages: containerImages,
	}
	var deployment *appsv1.Deployment

	deployment, err = templateParser.ParseResource[appsv1.Deployment](deploymentTemplatePath, deploymentConfig, reflect.TypeOf(&appsv1.Deployment{}))
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to parse deployment template")
	}
	if err := controllerutil.SetControllerReference(nemoGuardrails, deployment, r.Scheme); err != nil {
		log.FromContext(ctx).Error(err, "Failed to set controller reference for deployment")
	}
	return deployment
}
