package utils

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RouteConfig struct {
	ServiceName string
	Name        *string
	Namespace   *string
	Owner       *metav1.Object
	PortName    string
	Termination *string
}

const (
	Edge        = "edge"
	Reencrypt   = "reencrypt"
	Passthrough = "passthrough"
)

const routeResourceKind = "route"

// === GENERIC FUNCTIONS ===============================================================================================

// DefineRoute creates a route object, but does not deploy it to the cluster
func DefineRoute(ctx context.Context, c client.Client, owner metav1.Object, routeConfig RouteConfig, routeTemplatePath string, parser ResourceParserFunc[*routev1.Route]) (*routev1.Route, error) {
	// Create route object
	genericConfig := processRouteConfig(owner, routeConfig)
	return DefineGeneric[*routev1.Route](ctx, c, owner, routeResourceKind, genericConfig, routeTemplatePath, parser)
}

// ReconcileRoute contains logic for generic route reconciliation
func ReconcileRoute(ctx context.Context, c client.Client, owner metav1.Object, routeConfig RouteConfig, routeTemplatePath string, parserFunc ResourceParserFunc[*routev1.Route]) error {
	genericConfig := processRouteConfig(owner, routeConfig)
	_, _, err := ReconcileGeneric[*routev1.Route](ctx, c, owner, routeResourceKind, genericConfig, routeTemplatePath, parserFunc)
	return err
}

// === SPECIFIC ROUTE FUNCTIONS ========================================================================================

// processConfig sets default values for the RouteConfig if not provided, then converts into a GenericConfig
func processRouteConfig(owner metav1.Object, routeConfig RouteConfig) GenericConfig {
	if routeConfig.Name == nil {
		routeConfig.Name = StringPointer(owner.GetName())
	}

	if routeConfig.Namespace == nil {
		routeConfig.Namespace = StringPointer(owner.GetNamespace())
	}

	if routeConfig.Owner == nil {
		routeConfig.Owner = &owner
	}

	return GetGenericConfig(routeConfig.Name, routeConfig.Namespace, routeConfig)
}

// RouteIsAdmitted returns true if the route has an ingress condition RouteAdmitted=True.
func RouteIsAdmitted(r *routev1.Route) bool {
	if r == nil {
		return false
	}
	for _, ingress := range r.Status.Ingress {
		for _, condition := range ingress.Conditions {
			if condition.Type == routev1.RouteAdmitted && condition.Status == "True" {
				return true
			}
		}
	}
	return false
}

// serviceNameAndNamespaceFromHost parses a Kubernetes Service DNS host or short name.
// Examples: "my-svc" with defaultNS → my-svc/defaultNS; "my-svc.my-ns.svc.cluster.local" → my-svc/my-ns.
func serviceNameAndNamespaceFromHost(host, defaultNamespace string) (svcName, ns string) {
	h := strings.TrimSpace(strings.ToLower(host))
	h = strings.TrimSuffix(h, ".")
	if strings.HasSuffix(h, ".svc.cluster.local") {
		h = strings.TrimSuffix(h, ".svc.cluster.local")
	} else if strings.HasSuffix(h, ".svc") {
		h = strings.TrimSuffix(h, ".svc")
	}
	if h == "" {
		return "", defaultNamespace
	}
	if strings.Count(h, ".") == 1 {
		i := strings.Index(h, ".")
		return h[:i], h[i+1:]
	}
	if strings.Contains(h, ".") {
		return "", defaultNamespace
	}
	return h, defaultNamespace
}

// serviceMCPBackendReady checks that a Service exists and its Endpoints have at least one ready address.
func serviceMCPBackendReady(ctx context.Context, c client.Client, svcName, ns string) error {
	svc := &corev1.Service{}
	if err := c.Get(ctx, types.NamespacedName{Name: svcName, Namespace: ns}, svc); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("no OpenShift Route for host %q and no Service %q in namespace %q", svcName, svcName, ns)
		}
		return fmt.Errorf("get Service %q/%q: %w", ns, svcName, err)
	}
	eps := &corev1.Endpoints{}
	if err := c.Get(ctx, types.NamespacedName{Name: svcName, Namespace: ns}, eps); err != nil {
		return fmt.Errorf("get Endpoints for Service %q/%q: %w", ns, svcName, err)
	}
	for _, sub := range eps.Subsets {
		if len(sub.Addresses) > 0 {
			return nil
		}
	}
	return fmt.Errorf("Service %q/%q has no ready endpoints yet", ns, svcName)
}

// ValidateMCPURLUsesAdmittedRoute checks mcp.url from plugin config.
// If an OpenShift Route lists the same spec.host as the URL, that Route must be admitted.
// If no Route matches (typical in-cluster Service URL), the host is resolved as a Kubernetes Service
// in lookupNamespace (or service.namespace from *.svc DNS) and must have ready Endpoints.
func ValidateMCPURLUsesAdmittedRoute(ctx context.Context, c client.Client, rawURL string, lookupNamespace string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse mcp.url: %w", err)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("mcp.url has no host: %q", rawURL)
	}
	var list routev1.RouteList
	if err := c.List(ctx, &list); err != nil {
		return fmt.Errorf("list Routes: %w", err)
	}
	var routeMatchingHost *routev1.Route
	for i := range list.Items {
		r := &list.Items[i]
		if !strings.EqualFold(strings.TrimSpace(r.Spec.Host), host) {
			continue
		}
		if RouteIsAdmitted(r) {
			return nil
		}
		routeMatchingHost = r
	}
	if routeMatchingHost != nil {
		return fmt.Errorf("OpenShift Route %s/%s for host %q is not admitted yet", routeMatchingHost.Namespace, routeMatchingHost.Name, host)
	}
	svcName, ns := serviceNameAndNamespaceFromHost(host, lookupNamespace)
	if svcName == "" {
		return fmt.Errorf("no OpenShift Route with host %q and host is not a resolvable Kubernetes Service name", host)
	}
	return serviceMCPBackendReady(ctx, c, svcName, ns)
}

// CheckRouteReady verifies if a route is created and admitted
func CheckRouteReady(ctx context.Context, c client.Client, name string, namespace string) (bool, error) {
	// Retry logic for getting the route and checking its readiness
	var existingRoute *routev1.Route
	err := retry.OnError(
		wait.Backoff{
			Steps:    5,
			Duration: time.Second * 5,
			Factor:   1.0,
		},
		func(err error) bool {
			// Retry on transient errors, such as network errors or resource not found
			return errors.IsNotFound(err)
		},
		func() error {
			// Fetch the Route resource
			typedNamespaceName := types.NamespacedName{Name: name, Namespace: namespace}
			existingRoute = &routev1.Route{}
			err := c.Get(ctx, typedNamespaceName, existingRoute)
			if err != nil {
				return err
			}
			if RouteIsAdmitted(existingRoute) {
				return nil
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
