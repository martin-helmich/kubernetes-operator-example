package helloworld

import (
	"context"

	examplev1alpha1 "github.com/martin-helmich/helloworld-operator/pkg/apis/example/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_helloworld")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new HelloWorld Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileHelloWorld{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("helloworld-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource HelloWorld
	err = c.Watch(&source.Kind{Type: &examplev1alpha1.HelloWorld{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	watchTypes := []runtime.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&extv1beta1.Ingress{},
	}

	for i := range watchTypes {
		err = c.Watch(&source.Kind{Type: watchTypes[i]}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &examplev1alpha1.HelloWorld{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileHelloWorld{}

// ReconcileHelloWorld reconciles a HelloWorld object
type ReconcileHelloWorld struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a HelloWorld object and makes changes based on the state read
// and what is in the HelloWorld.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHelloWorld) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling HelloWorld")

	ctx := context.TODO()

	// Fetch the HelloWorld instance
	instance := &examplev1alpha1.HelloWorld{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	labels := map[string]string{
		"app": instance.Name,
	}

	deployment, err := r.buildDeployment(instance, labels)
	if err != nil {
		return reconcile.Result{}, err
	}

	service, err := r.buildService(instance, labels)
	if err != nil {
		return reconcile.Result{}, err
	}

	ingress, err := r.buildIngress(instance, labels)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Deployment already exists
	foundDepl := appsv1.Deployment{}
	foundService := corev1.Service{}
	foundIngress := extv1beta1.Ingress{}

	err = r.client.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, &foundDepl)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		if err := r.client.Create(ctx, deployment); err != nil {
			return reconcile.Result{}, err
		}
	} else if err == nil && foundDepl.Spec.Replicas != instance.Spec.Replicas {
		reqLogger.Info("updating existing Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		foundDepl.Spec.Replicas = instance.Spec.Replicas
		if err := r.client.Update(ctx, &foundDepl); err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	err = r.client.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &foundService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		if err := r.client.Create(ctx, service); err != nil {
			return reconcile.Result{}, err
		}
	} else if err == nil {
		reqLogger.Info("updating existing Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)

		foundService.Spec.Ports = service.Spec.Ports

		if err := r.client.Update(ctx, &foundService); err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	err = r.client.Get(ctx, types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, &foundIngress)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("creating a new Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
		if err := r.client.Create(ctx, ingress); err != nil {
			return reconcile.Result{}, err
		}
	} else if err == nil {
		reqLogger.Info("updating existing Ingress", "Ingress.Namespace", ingress.Namespace, "Ingress.Name", ingress.Name)
		ingress.Spec.DeepCopyInto(&foundIngress.Spec)
		if err := r.client.Update(ctx, &foundIngress); err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileHelloWorld) buildDeployment(cr *examplev1alpha1.HelloWorld, labels map[string]string) (*appsv1.Deployment, error) {
	recipient := cr.Spec.Recipient
	if recipient == "" {
		recipient = "World"
	}

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cr.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "helloworld",
							Image: "martinhelmich/helloworld",
							Env: []corev1.EnvVar{
								corev1.EnvVar{Name: "HELLOWORLD_GREETING", Value: recipient},
							},
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{Name: "http", ContainerPort: 8080},
							},
						},
					},
				},
			},
		},
	}

	// Set HelloWorld instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, &deployment, r.scheme); err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (r *ReconcileHelloWorld) buildService(cr *examplev1alpha1.HelloWorld, labels map[string]string) (*corev1.Service, error) {
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{Name: "http", TargetPort: intstr.FromString("http"), Port: 80},
			},
			Selector: labels,
		},
	}

	// Set HelloWorld instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, &service, r.scheme); err != nil {
		return nil, err
	}

	return &service, nil
}

func (r *ReconcileHelloWorld) buildIngress(cr *examplev1alpha1.HelloWorld, labels map[string]string) (*extv1beta1.Ingress, error) {
	ingress := extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: extv1beta1.IngressSpec{
			Rules: []extv1beta1.IngressRule{
				extv1beta1.IngressRule{
					Host: cr.Spec.Host,
					IngressRuleValue: extv1beta1.IngressRuleValue{
						HTTP: &extv1beta1.HTTPIngressRuleValue{
							Paths: []extv1beta1.HTTPIngressPath{
								extv1beta1.HTTPIngressPath{
									Path: "/",
									Backend: extv1beta1.IngressBackend{
										ServiceName: cr.Name,
										ServicePort: intstr.FromString("http"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(cr, &ingress, r.scheme); err != nil {
		return nil, err
	}

	return &ingress, nil
}
