package utils

import (
	"context"
	templateParser "github.com/trustyai-explainability/trustyai-service-operator/controllers/tas/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ServiceConfig struct {
	Owner *metav1.Object
}

func CreateService(ctx context.Context, c client.Client, owner metav1.Object, templatePath string) *corev1.Service {
	serviceConfig := ServiceConfig{
		Owner: &owner,
	}
	var service *corev1.Service
	service, err := templateParser.ParseResource[corev1.Service](templatePath, serviceConfig, reflect.TypeOf(&corev1.Service{}))
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to parse service template")
	}
	err = controllerutil.SetControllerReference(owner, service, c.Scheme())
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to set controller reference")
		return nil
	}
	return service
}
