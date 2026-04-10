package utils

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	envoyFilterResourceKind = "EnvoyFilter"
	envoyFilterAPIVersion   = "networking.istio.io/v1alpha3"
)

type EnvoyFilterConfig struct {
	Name           string
	Namespace      string
	OwnerName      string
	OwnerNamespace string
	GatewayName    string
	ServiceAddress string
	GRPCPort       int
	PortNumber     int
}

func GetEnvoyFilterName(name string) string {
	return name + "-envoyfilter"
}

// ReconcileEnvoyFilter creates or updates an EnvoyFilter
func ReconcileEnvoyFilter(ctx context.Context, c client.Client, config EnvoyFilterConfig, templatePath string, parser ResourceParserFunc[*unstructured.Unstructured]) error {
	logger := log.FromContext(ctx)

	desired, err := parser(templatePath, config, reflect.TypeOf(&unstructured.Unstructured{}))
	if err != nil {
		LogErrorParsing(ctx, err, envoyFilterResourceKind, config.Name, config.Namespace)
		return err
	}

	existing := &unstructured.Unstructured{}
	existing.SetAPIVersion(envoyFilterAPIVersion)
	existing.SetKind(envoyFilterResourceKind)

	err = c.Get(ctx, types.NamespacedName{Name: config.Name, Namespace: config.Namespace}, existing)
	if errors.IsNotFound(err) {
		LogInfoCreating(ctx, envoyFilterResourceKind, config.Name, config.Namespace)
		if err := c.Create(ctx, desired); err != nil {
			LogErrorCreating(ctx, err, envoyFilterResourceKind, config.Name, config.Namespace)
			return err
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check for existing EnvoyFilter %s/%s: %w", config.Namespace, config.Name, err)
	}

	if !equality.Semantic.DeepEqual(existing.Object["spec"], desired.Object["spec"]) {
		logger.Info("Updating EnvoyFilter", "name", config.Name, "namespace", config.Namespace)
		existing.Object["spec"] = desired.Object["spec"]
		existing.SetLabels(desired.GetLabels())
		if err := c.Update(ctx, existing); err != nil {
			LogErrorUpdating(ctx, err, envoyFilterResourceKind, config.Name, config.Namespace)
			return err
		}
	}
	return nil
}

func DeleteEnvoyFiltersByOwner(ctx context.Context, c client.Client, ownerName, ownerNamespace string) error {
	list := &unstructured.UnstructuredList{}
	list.SetAPIVersion(envoyFilterAPIVersion)
	list.SetKind(envoyFilterResourceKind)

	if err := c.List(ctx, list, client.MatchingLabels{
		"app": ownerName,
		"trustyai.opendatahub.io/owner-namespace": ownerNamespace,
		"app.kubernetes.io/managed-by":            "trustyai-service-operator",
	}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("list EnvoyFilters owned by %s/%s: %w", ownerNamespace, ownerName, err)
	}

	for i := range list.Items {
		log.FromContext(ctx).Info("Deleting EnvoyFilter",
			"name", list.Items[i].GetName(), "namespace", list.Items[i].GetNamespace())
		if err := c.Delete(ctx, &list.Items[i]); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("delete EnvoyFilter %s/%s: %w",
				list.Items[i].GetNamespace(), list.Items[i].GetName(), err)
		}
	}
	return nil
}
