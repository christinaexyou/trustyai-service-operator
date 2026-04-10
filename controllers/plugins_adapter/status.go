package plugins_adapter

import (
	"context"

	routev1 "github.com/openshift/api/route/v1"
	pluginsadapterv1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/plugins_adapter/v1alpha1"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *PluginsAdapterReconciler) updateStatus(ctx context.Context, original *pluginsadapterv1alpha1.PluginsAdapter, update func(saved *pluginsadapterv1alpha1.PluginsAdapter)) (*pluginsadapterv1alpha1.PluginsAdapter, error) {
	saved := original.DeepCopy()

	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if err := r.Client.Get(ctx, client.ObjectKeyFromObject(original), saved); err != nil {
			return err
		}
		update(saved)
		return r.Client.Status().Update(ctx, saved)
	})
	return saved, err
}

func (r *PluginsAdapterReconciler) isDeploymentReady(ctx context.Context, name, namespace string) bool {
	deployment := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment); err != nil {
		return false
	}
	for _, cond := range deployment.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *PluginsAdapterReconciler) isRouteAdmitted(ctx context.Context, name, namespace string) bool {
	route := &routev1.Route{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, route); err != nil {
		if !errors.IsNotFound(err) {
			log.FromContext(ctx).Error(err, "Failed to get Route", "name", name, "namespace", namespace)
		}
		return false
	}
	return utils.RouteIsAdmitted(route)
}

func (r *PluginsAdapterReconciler) reconcileStatuses(ctx context.Context, pluginsAdapter *pluginsadapterv1alpha1.PluginsAdapter) (ctrl.Result, error) {
	deploymentReady := r.isDeploymentReady(ctx, pluginsAdapter.Name, pluginsAdapter.Namespace)
	routeReady := r.isRouteAdmitted(ctx, pluginsAdapter.Name, pluginsAdapter.Namespace)

	if deploymentReady && routeReady {
		_, updateErr := r.updateStatus(ctx, pluginsAdapter, func(saved *pluginsadapterv1alpha1.PluginsAdapter) {
			utils.SetResourceCondition(&saved.Status.Conditions, "Deployment", "DeploymentReady", "Deployment is ready", corev1.ConditionTrue)
			utils.SetResourceCondition(&saved.Status.Conditions, "Route", "RouteReady", "Route is ready", corev1.ConditionTrue)
			utils.SetCompleteCondition(&saved.Status.Conditions, corev1.ConditionTrue, utils.ReconcileCompleted, utils.ReconcileCompletedMessage)
			saved.Status.Phase = utils.PhaseReady
		})
		if updateErr != nil {
			log.FromContext(ctx).Error(updateErr, "Failed to update status")
			return ctrl.Result{}, updateErr
		}
		return ctrl.Result{}, nil
	}

	_, updateErr := r.updateStatus(ctx, pluginsAdapter, func(saved *pluginsadapterv1alpha1.PluginsAdapter) {
		utils.SetStatus(&saved.Status.Conditions, "Deployment", deploymentReady)
		utils.SetStatus(&saved.Status.Conditions, "Route", routeReady)
		utils.SetCompleteCondition(&saved.Status.Conditions, corev1.ConditionFalse, utils.ReconcileFailed, utils.ReconcileFailedMessage)
		saved.Status.Phase = "Progressing"
	})
	if updateErr != nil {
		log.FromContext(ctx).Error(updateErr, "Failed to update status")
		return ctrl.Result{}, updateErr
	}
	return ctrl.Result{Requeue: true}, nil
}
