package guardrails

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	guardrailsv1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/guardrails/v1alpha1"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// The registered function to set up the Guardrails controller
func ControllerSetUp(mgr manager.Manager, ns, configmap string, recorder record.EventRecorder) error {
	return (&GuardrailsReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Namespace:     ns,
		EventRecorder: recorder,
	}).SetupWithManager(mgr)
}

// GuardrailsReconciler reconciles a GuardrailsOrchestrator object
type GuardrailsReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Namespace     string
	EventRecorder record.EventRecorder
}

func (r *GuardrailsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Fetch the GuardrailsOrchestrator instance
	orchestrator := &guardrailsv1alpha1.GuardrailsOrchestrator{}
	err := r.Get(ctx, req.NamespacedName, orchestrator)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("GuardrailsOrchestrator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get GuardrailsOrchestrator")
		return ctrl.Result{}, err
	}

	// Add the finalizer if it does not exist
	if !utils.ContainsString(orchestrator.Finalizers, FinalizerName) {
		log.Info("Adding finalizer to GuardrailsOrchestrator", "finalizer", FinalizerName)
		orchestrator.Finalizers = append(orchestrator.Finalizers, FinalizerName)
		if err := r.Update(ctx, orchestrator); err != nil {
			log.Error(err, "Failed to add the finalizer to GuardrailsOrchestrator")
		}
	}

	// Handle deletion
	if !orchestrator.DeletionTimestamp.IsZero() {
		if utils.ContainsString(orchestrator.Finalizers, FinalizerName) {
			// Delete the deployment
			if err := r.deleteDeployment(ctx, orchestrator); err != nil {
				log.Error(err, "unable to delete GuardrailsOrchestrator deployment")
			}
			log.Info("deleted GuardrailsOrchestrator deployment")
		}
		orchestrator.Finalizers = utils.RemoveString(orchestrator.Finalizers, FinalizerName)
		if err := r.Update(ctx, orchestrator); err != nil {
			log.Error(err, "failed to remove the finalizer")
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Create the service account
	err = r.ensureServiceAccount(ctx, orchestrator)
	if err != nil {
		log.Error(err, "ServiceAccount not ready")
		return ctrl.Result{Requeue: true}, err
	}

	// Create external routes for inference services
	shouldContinue, err := r.ensureInferenceServices(ctx, orchestrator, orchestrator.Namespace)
	if err != nil {
		log.Error(err, "Failed to ensure inference services")
		return ctrl.Result{Requeue: true}, err
	}
	if !shouldContinue {
		log.Error(err, "Failed to ")
		return ctrl.Result{Requeue: true}, err
	}

	// Create the deployment
	err = r.ensureDeployment(ctx, orchestrator)
	if err != nil {
		log.Error(err, "Deployment not ready")
		return ctrl.Result{}, err
	}

	// Check if the service already exists, if not create the service
	err = r.ensureService(ctx, orchestrator)
	if err != nil {
		log.Error(err, "Service not ready")
		return ctrl.Result{}, err
	}

	// Check if the Route already exists, if not create the route
	err = r.ensureRoute(ctx, orchestrator)
	if err != nil {
		log.Error(err, "Route not ready")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GuardrailsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&guardrailsv1alpha1.GuardrailsOrchestrator{}).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestForOwner{
				OwnerType:    &guardrailsv1alpha1.GuardrailsOrchestrator{},
				IsController: true,
			},
		).
		Complete(r)
}