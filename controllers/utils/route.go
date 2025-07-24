package utils

import (
	"context"
	"fmt"
	routev1 "github.com/openshift/api/route/v1"
	templateParser "github.com/trustyai-explainability/trustyai-service-operator/controllers/tas/templates"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

type RouteConfig struct {
	Owner metav1.Object
}

func createRoute(ctx context.Context, c client.Client, owner metav1.Object, routeTemplatePath string) *routev1.Route {
	routeHttpsConfig := RouteConfig{
		Owner: owner,
	}
	var route *routev1.Route
	route, err := templateParser.ParseResource[routev1.Route](routeTemplatePath, routeHttpsConfig, reflect.TypeOf(&routev1.Route{}))

	if err != nil {
		log.FromContext(ctx).Error(err, "failed to parse route template")
	}
	err = controllerutil.SetControllerReference(owner, route, c.Scheme())
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to set controller reference")
		return nil
	}
	return route
}

func CheckRouteReady(ctx context.Context, c client.Client, name string, namespace string, portName string) (bool, error) {
	// Retry logic for getting the route and checking its readiness
	var existingRoute *routev1.Route
	err := retry.OnError(
		wait.Backoff{
			Duration: time.Second * 5,
		},
		func(err error) bool {
			// Retry on transient errors, such as network errors or resource not found
			return errors.IsNotFound(err) || err != nil
		},
		func() error {
			// Fetch the Route resource
			typedNamespaceName := types.NamespacedName{Name: name + portName, Namespace: namespace}
			existingRoute = &routev1.Route{}
			err := c.Get(ctx, typedNamespaceName, existingRoute)
			if err != nil {
				return err
			}

			for _, ingress := range existingRoute.Status.Ingress {
				for _, condition := range ingress.Conditions {
					if condition.Type == routev1.RouteAdmitted && condition.Status == "True" {
						return nil
					}
				}
			}
			// Route is not admitted yet, return an error to retry
			return fmt.Errorf("route %s is not admitted", name)
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}
