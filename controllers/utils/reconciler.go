package utils

import (
	"context"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileService(ctx context.Context, c client.Client, owner metav1.Object, templatePath string, log logr.Logger) (ctrl.Result, error) {
	existingService := &corev1.Service{}
	err := c.Get(ctx, types.NamespacedName{Name: owner.GetName() + "-service", Namespace: owner.GetNamespace()}, existingService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		service := CreateService(ctx, c, owner, templatePath)
		log.Info("Creating a new Service", "Service.Namespace", owner.GetNamespace(), "Service.Name", owner.GetName())
		err = c.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", owner.GetNamespace(), "Service.Name", owner.GetName())
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func ReconcileRoute(ctx context.Context, c client.Client, owner metav1.Object, templatePath string, log logr.Logger) (ctrl.Result, error) {
	existingRoute := &routev1.Route{}
	err := c.Get(ctx, types.NamespacedName{Name: owner.GetName(), Namespace: owner.GetNamespace()}, existingRoute)
	if err != nil && errors.IsNotFound(err) {
		// Define a new route
		route := createRoute(ctx, c, owner, templatePath)
		log.Info("Creating a new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
		err = c.Create(ctx, route)
		if err != nil {
			log.Error(err, "Failed to create new Route", "Route.Namespace", route.Namespace, "Route.Name", route.Name)
		}
	} else if err != nil {
		log.Error(err, "Failed to get Route")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
