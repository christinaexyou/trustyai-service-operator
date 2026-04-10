package plugins_adapter

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pluginsadapterv1alpha1 "github.com/trustyai-explainability/trustyai-service-operator/api/plugins_adapter/v1alpha1"
	"github.com/trustyai-explainability/trustyai-service-operator/controllers/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("PluginsAdapter Controller", func() {
	const (
		resourceName      = "test-pluginsadapter"
		namespace         = "test-gateway-ns"
		operatorNamespace = "operator-ns"
	)

	var (
		ctx                 = context.Background()
		typedNamespacedName = types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
	)

	BeforeEach(func() {
		By("creating the test namespace")
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}
		err := k8sClient.Create(ctx, ns)
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		By("creating the operator namespace")
		opNs := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: operatorNamespace},
		}
		err = k8sClient.Create(ctx, opNs)
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		By("creating the operator config ConfigMap")
		operatorConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ConfigMap,
				Namespace: operatorNamespace,
			},
			Data: map[string]string{
				pluginsAdapterImageKey: "quay.io/trustyai/plugins-adapter:latest",
			},
		}
		err = k8sClient.Create(ctx, operatorConfigMap)
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		By("creating the plugins-adapter-config ConfigMap")
		pluginsAdapterConfigMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "plugins-adapter-config",
				Namespace: namespace,
			},
			Data: map[string]string{
				"config.yaml": "dummy: config",
			},
		}
		err = k8sClient.Create(ctx, pluginsAdapterConfigMap)
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		By("creating the PluginsAdapter CR")
		pa := &pluginsadapterv1alpha1.PluginsAdapter{}
		err = k8sClient.Get(ctx, typedNamespacedName, pa)
		if err != nil && errors.IsNotFound(err) {
			log.FromContext(ctx).Info("Creating a new PluginsAdapter resource")
			resource := &pluginsadapterv1alpha1.PluginsAdapter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: pluginsadapterv1alpha1.PluginsAdapterSpec{
					GatewayConfig: &pluginsadapterv1alpha1.GatewayConfig{
						Name:      "test-mcp-gateway",
						Namespace: namespace,
					},
					PluginsConfigs: []pluginsadapterv1alpha1.PluginConfig{
						{
							Name:       "default-config",
							ConfigMaps: []string{"plugins-adapter-config"},
							Default:    true,
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		}
	})

	AfterEach(func() {
		resource := &pluginsadapterv1alpha1.PluginsAdapter{}
		err := k8sClient.Get(ctx, typedNamespacedName, resource)
		if err == nil {
			By("Cleaning up the PluginsAdapter resource")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			controllerReconciler := &PluginsAdapterReconciler{
				Client:    k8sClient,
				Scheme:    k8sClient.Scheme(),
				Namespace: operatorNamespace,
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typedNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, typedNamespacedName, resource)
				return errors.IsNotFound(err)
			}, time.Second*5, time.Millisecond*100).Should(BeTrue())
		}
	})

	It("should successfully reconcile the resource and create a Deployment", func() {
		By("Reconciling the created resource")
		controllerReconciler := &PluginsAdapterReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			Namespace: operatorNamespace,
		}
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Checking if the Deployment was created")
		Eventually(func() error {
			deployment := &appsv1.Deployment{}
			return k8sClient.Get(ctx, typedNamespacedName, deployment)
		}, time.Second*10, time.Millisecond*100).Should(Succeed())
	})

	It("should set the CONFIG_ID env var to the default config", func() {
		By("Reconciling the created resource")
		controllerReconciler := &PluginsAdapterReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			Namespace: operatorNamespace,
		}
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Checking if the Deployment contains the CONFIG_ID environment variable")
		Eventually(func() bool {
			deployment := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, typedNamespacedName, deployment)
			if err != nil {
				return false
			}
			for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
				if env.Name == "CONFIG_ID" && env.Value == "default-config" {
					return true
				}
			}
			return false
		}, time.Second*10, time.Millisecond*100).Should(BeTrue())
	})

	It("should add user environment variables to the deployment", func() {
		By("Updating the PluginsAdapter CR with user env vars")
		pa := &pluginsadapterv1alpha1.PluginsAdapter{}
		Expect(k8sClient.Get(ctx, typedNamespacedName, pa)).To(Succeed())
		pa.Spec.Env = []corev1.EnvVar{
			{Name: "CUSTOM_VAR", Value: "custom-value"},
		}
		Expect(k8sClient.Update(ctx, pa)).To(Succeed())

		By("Reconciling the resource")
		controllerReconciler := &PluginsAdapterReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			Namespace: operatorNamespace,
		}
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Checking if the Deployment contains the user-provided environment variable")
		Eventually(func() bool {
			deployment := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, typedNamespacedName, deployment)
			if err != nil {
				return false
			}
			for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
				if env.Name == "CUSTOM_VAR" && env.Value == "custom-value" {
					return true
				}
			}
			return false
		}, time.Second*10, time.Millisecond*100).Should(BeTrue())
	})

	It("should create a Service", func() {
		By("Reconciling the created resource")
		controllerReconciler := &PluginsAdapterReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			Namespace: operatorNamespace,
		}
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Checking if the Service was created")
		Eventually(func() error {
			service := &corev1.Service{}
			return k8sClient.Get(ctx, typedNamespacedName, service)
		}, time.Second*10, time.Millisecond*100).Should(Succeed())
	})

	It("should mount plugin config volumes on the deployment", func() {
		By("Reconciling the created resource")
		controllerReconciler := &PluginsAdapterReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			Namespace: operatorNamespace,
		}
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Checking that the ConfigMap volume is mounted")
		Eventually(func() bool {
			deployment := &appsv1.Deployment{}
			err := k8sClient.Get(ctx, typedNamespacedName, deployment)
			if err != nil {
				return false
			}
			for _, vm := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
				if vm.MountPath == "/app/config/default-config/0" {
					return true
				}
			}
			return false
		}, time.Second*10, time.Millisecond*100).Should(BeTrue())
	})

	It("should handle deletion with finalizer cleanup", func() {
		controllerReconciler := &PluginsAdapterReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			Namespace: operatorNamespace,
		}

		By("Reconciling to add the finalizer")
		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Verifying the finalizer was added")
		pa := &pluginsadapterv1alpha1.PluginsAdapter{}
		Expect(k8sClient.Get(ctx, typedNamespacedName, pa)).To(Succeed())
		Expect(pa.Finalizers).To(ContainElement(finalizerName))

		By("Deleting the resource")
		Expect(k8sClient.Delete(ctx, pa)).To(Succeed())

		By("Reconciling after deletion to remove the finalizer")
		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: typedNamespacedName,
		})
		Expect(err).NotTo(HaveOccurred())

		By("Verifying the resource is fully deleted")
		Eventually(func() bool {
			err := k8sClient.Get(ctx, typedNamespacedName, &pluginsadapterv1alpha1.PluginsAdapter{})
			return errors.IsNotFound(err)
		}, time.Second*5, time.Millisecond*100).Should(BeTrue())
	})
})
