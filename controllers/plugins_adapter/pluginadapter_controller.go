package plugins_adapter

import (
	"context"
	"fmt"
	"time"

	pluginsadapterv1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/plugins_adapter/v1alpha1"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/constants"
	templateParser "github.com/trustyai-explainability/trustyai-service-operator/controllers/plugins_adapter/templates"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// PluginsAdapterReconciler reconciles a PluginsAdapter object
type PluginsAdapterReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Namespace string
	Recorder  record.EventRecorder
}

const (
	serviceTemplate     = "service.tmpl.yaml"
	routeTemplate       = "route.tmpl.yaml"
	envoyFilterTemplate = "envoy-filter.tmpl.yaml"
)

//+kubebuilder:rbac:groups=trustyai.opendatahub.io,resources=pluginsadapters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=trustyai.opendatahub.io,resources=pluginsadapters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=trustyai.opendatahub.io,resources=pluginsadapters/finalizers,verbs=update
//+kubebuilder:rbac:groups=trustyai.opendatahub.io,resources=nemoguardrails,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=list;watch;get;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.istio.io,resources=envoyfilters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *PluginsAdapterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// ====== Fetch instance of PluginsAdapter CR ======================================================================
	pluginsAdapter := &pluginsadapterv1alpha1.PluginsAdapter{}
	err := r.Get(ctx, req.NamespacedName, pluginsAdapter)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("PluginsAdapter resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		utils.LogErrorRetrieving(ctx, err, "PluginsAdapter Custom Resource", pluginsAdapter.Name, pluginsAdapter.Namespace)
		return ctrl.Result{}, err
	}

	// ====== Deletion handling ========================================================================================
	if err := utils.AddFinalizerIfNeeded(ctx, r.Client, pluginsAdapter, finalizerName); err != nil {
		return ctrl.Result{}, err
	}

	// define all the cleanup steps needed before the finalizer can be removed in the CleanupFunc
	cleanupFunc := func() error {
		return utils.CleanupClusterRoleBinding(ctx, r.Client, pluginsAdapter)
	}
	shouldExit, err := utils.HandleDeletionIfNeeded(ctx, r.Client, pluginsAdapter, finalizerName, cleanupFunc)
	if err != nil {
		return ctrl.Result{}, err
	}
	if shouldExit {
		return ctrl.Result{}, nil
	}

	// ====== Reconcile Deployment =====================================================================================
	existingDeployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: pluginsAdapter.Name, Namespace: pluginsAdapter.Namespace}, existingDeployment)
	if err != nil && errors.IsNotFound(err) {
		deployment, err := r.createDeployment(ctx, pluginsAdapter)
		if err != nil {
			return ctrl.Result{}, err
		}
		utils.LogInfoCreating(ctx, "deployment", pluginsAdapter.Name, pluginsAdapter.Namespace)
		if err = r.Create(ctx, deployment); err != nil {
			utils.LogErrorCreating(ctx, err, "deployment", deployment.Name, deployment.Namespace)
			return ctrl.Result{}, err
		}
	} else if err != nil {
		utils.LogErrorRetrieving(ctx, err, "deployment", pluginsAdapter.Name, pluginsAdapter.Namespace)
		return ctrl.Result{}, err
	} else {
		deployment, err := r.createDeployment(ctx, pluginsAdapter)
		if err != nil {
			return ctrl.Result{}, err
		}
		if !equality.Semantic.DeepEqual(existingDeployment.Spec, deployment.Spec) {
			existingDeployment.Spec = deployment.Spec
			if err = r.Update(ctx, existingDeployment); err != nil {
				utils.LogErrorUpdating(ctx, err, "deployment", existingDeployment.Name, existingDeployment.Namespace)
				return ctrl.Result{}, err
			}
		}
	}

	// ====== Reconcile Service ========================================================================================
	serviceConfig := utils.ServiceConfig{
		Name:         pluginsAdapter.Name,
		Namespace:    pluginsAdapter.Namespace,
		Owner:        pluginsAdapter,
		Version:      constants.Version,
		UseAuthProxy: utils.RequiresAuth(pluginsAdapter),
	}
	if err = utils.ReconcileService(ctx, r.Client, pluginsAdapter, serviceConfig, serviceTemplate, templateParser.ParseResource); err != nil {
		utils.LogErrorReconciling(ctx, err, "service", serviceConfig.Name, serviceConfig.Namespace)
		return ctrl.Result{}, err
	}

	// ====== Reconcile Route ==========================================================================================
	routeConfig := utils.RouteConfig{
		PortName:    pluginsAdapter.Name,
		ServiceName: pluginsAdapter.Name,
		Termination: utils.StringPointer(utils.Edge),
	}
	if err = utils.ReconcileRoute(ctx, r.Client, pluginsAdapter, routeConfig, routeTemplate, templateParser.ParseResource); err != nil {
		utils.LogErrorReconciling(ctx, err, "route", pluginsAdapter.Name, pluginsAdapter.Namespace)
		return ctrl.Result{}, err
	}

	// ====== Reconcile Envoy Filter ===================================================================================
	envoyFilterName := utils.GetEnvoyFilterName(pluginsAdapter.Name)
	envoyFilterNS := pluginsAdapter.Namespace
	gatewayName := "mcp-gateway"
	if pluginsAdapter.Spec.GatewayConfig != nil {
		if pluginsAdapter.Spec.GatewayConfig.Namespace != "" {
			envoyFilterNS = pluginsAdapter.Spec.GatewayConfig.Namespace
		}
		if pluginsAdapter.Spec.GatewayConfig.Name != "" {
			gatewayName = pluginsAdapter.Spec.GatewayConfig.Name
		}
	}
	envoyFilterConfig := utils.EnvoyFilterConfig{
		Name:           envoyFilterName,
		Namespace:      envoyFilterNS,
		OwnerName:      pluginsAdapter.Name,
		OwnerNamespace: pluginsAdapter.Namespace,
		GatewayName:    gatewayName,
		ServiceAddress: fmt.Sprintf("%s-service.%s.svc.cluster.local", pluginsAdapter.Name, pluginsAdapter.Namespace),
		GRPCPort:       50052,
		PortNumber:     8443,
	}
	if err = utils.ReconcileEnvoyFilter(ctx, r.Client, envoyFilterConfig, envoyFilterTemplate, templateParser.ParseResource); err != nil {
		utils.LogErrorReconciling(ctx, err, "envoy filter", envoyFilterName, envoyFilterNS)
		return ctrl.Result{}, err
	}

	// ====== Finalize reconciliation ==================================================================================
	result, updateErr := r.reconcileStatuses(ctx, pluginsAdapter)
	if updateErr != nil {
		return ctrl.Result{}, updateErr
	}
	if result.Requeue {
		logger.Info("Not all sub-resources are ready, requeuing")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	logger.Info("Reconciliation complete")
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PluginsAdapterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pluginsadapterv1alpha1.PluginsAdapter{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

// ControllerSetUp is the registered function to set up the PluginsAdapter controller.
func ControllerSetUp(mgr manager.Manager, ns, configmap string, recorder record.EventRecorder) error {
	return (&PluginsAdapterReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Namespace: ns,
		Recorder:  recorder,
	}).SetupWithManager(mgr)
}
