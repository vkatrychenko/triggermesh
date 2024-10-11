/*
Copyright 2022 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package reconciler

import (
	"context"

	"go.uber.org/zap"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacclientv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	appslistersv1 "k8s.io/client-go/listers/apps/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	rbaclistersv1 "k8s.io/client-go/listers/rbac/v1"
	"k8s.io/client-go/tools/cache"

	k8sclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformerv1 "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	replicasetinformerv1 "knative.dev/pkg/client/injection/kube/informers/apps/v1/replicaset"
	podinformerv1 "knative.dev/pkg/client/injection/kube/informers/core/v1/pod"
	sainformerv1 "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	rbinformerv1 "knative.dev/pkg/client/injection/kube/informers/rbac/v1/rolebinding"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/tracker"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingclientv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	servingclient "knative.dev/serving/pkg/client/injection/client"
	serviceinformerv1 "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"
	servinglistersv1 "knative.dev/serving/pkg/client/listers/serving/v1"
)

// GenericDeploymentReconciler contains interfaces shared across Deployment reconcilers.
type GenericDeploymentReconciler[T kmeta.OwnerRefable, L Lister[T]] struct {
	// URI resolver for sinks
	SinkResolver *resolver.URIResolver
	// API clients
	Client k8sClientGetter[*appsv1.Deployment, appsclientv1.DeploymentInterface]
	// objects listers
	Lister    func(namespace string) appslistersv1.DeploymentNamespaceLister
	PodLister func(namespace string) corelistersv1.PodNamespaceLister

	*GenericRBACReconciler[T, L]
}

// GenericServiceReconciler contains interfaces shared across Service reconcilers.
type GenericServiceReconciler[T kmeta.OwnerRefable, L Lister[T]] struct {
	// URI resolver for sinks
	SinkResolver *resolver.URIResolver
	// API clients
	Client k8sClientGetter[*servingv1.Service, servingclientv1.ServiceInterface]
	// objects listers
	Lister func(namespace string) servinglistersv1.ServiceNamespaceLister

	*GenericRBACReconciler[T, L]
}

// GenericRBACReconciler reconciles RBAC objects for components adapters.
type GenericRBACReconciler[T kmeta.OwnerRefable, L Lister[T]] struct {
	// API clients
	SAClient func(namespace string) coreclientv1.ServiceAccountInterface
	RBClient func(namespace string) rbacclientv1.RoleBindingInterface
	// objects listers
	SALister     func(namespace string) corelistersv1.ServiceAccountNamespaceLister
	RBLister     func(namespace string) rbaclistersv1.RoleBindingNamespaceLister
	OwnersLister ListerGetter[T, L]
}

// Lister is a partial generic version of a typed <Kind>NamespaceLister
// interface (usually generated by lister-gen).
type Lister[T kmeta.OwnerRefable] interface {
	List(labels.Selector) ([]T, error)
}

// ListerGetter obtains a namespaced Lister.
/*
  We deliberately use the signature
    func[T, L Lister[T]] func(namespace string) L
  instead of
    func[T] func(namespace string) Lister[T]

  Although both are functionally equivalent, the second form doesn't allow us to instantiate
  GenericRBACReconciler structs directly using the typed <Kind>NamespaceLister interfaces
  generated by lister-gen. Attempting to do so yields the following compiler error:

    cannot use informer.Lister().<Kind> (value of type func(namespace string) v1alpha1.<Kind>NamespaceLister)
      as type ListerGetter[*v1alpha1.<Kind>]

  Callers can circumvent this limitation by wrapping a typed lister getter inside a function
  to make it generic:

    func(namespace string) common.Lister[*v1alpha1.<Kind>] {
        return informer.Lister().<Kind>(namespace)
    }

  but this places the burden on the caller instead of on the compiler, which is suboptimal.

  With the first form, however, <Kind>NamespaceLister interfaces are handled as generic
  types without requiring an explicit conversion by the caller.
*/
type ListerGetter[T kmeta.OwnerRefable, L Lister[T]] func(namespace string) L

// k8sLister is like Lister, but suitable for core k8s objects, which do not
// implement kmeta.OwnerRefable.
type k8sLister[T metav1.Object] interface {
	List(labels.Selector) ([]T, error)
}

// k8sListerGetter is like ListerGetter, but suitable for core k8s objects,
// which do not implement kmeta.OwnerRefable.
type k8sListerGetter[T metav1.Object, L k8sLister[T]] func(namespace string) L

// k8sClient is a partial generic version of a typed <Kind>Interface client
// interface (usually generated by client-gen).
type k8sClient[T metav1.Object] interface {
	Create(context.Context, T, metav1.CreateOptions) (T, error)
	Update(context.Context, T, metav1.UpdateOptions) (T, error)
}

// k8sClientGetter obtains a namespaced k8sClient.
type k8sClientGetter[T metav1.Object, C k8sClient[T]] func(namespace string) C

// NewGenericDeploymentReconciler creates a new GenericDeploymentReconciler and
// attaches a default event handler to its Deployment informer.
func NewGenericDeploymentReconciler[T kmeta.OwnerRefable, L Lister[T]](ctx context.Context, gvk schema.GroupVersionKind,
	tracker tracker.Interface,
	adapterHandlerFn func(obj interface{}),
	ownersLister ListerGetter[T, L],
) GenericDeploymentReconciler[T, L] {

	deplInformer := deploymentinformerv1.Get(ctx)
	podInformer := podinformerv1.Get(ctx)

	r := GenericDeploymentReconciler[T, L]{
		SinkResolver:          resolver.NewURIResolverFromTracker(ctx, tracker),
		Client:                k8sclient.Get(ctx).AppsV1().Deployments,
		Lister:                deplInformer.Lister().Deployments,
		PodLister:             podInformer.Lister().Pods,
		GenericRBACReconciler: NewGenericRBACReconciler(ctx, ownersLister),
	}

	deplInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGVK(gvk),
		Handler:    controller.HandleAll(adapterHandlerFn),
	})

	var outermostCtlrType T

	podinformerv1.Get(ctx).Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: adapterPodWithAncestorOfKind(ctx, outermostCtlrType),
		Handler: controller.HandleAll(
			podOutermostAncestorHandlerFn(ctx, adapterHandlerFn),
		),
	})

	return r
}

// NewGenericServiceReconciler creates a new GenericServiceReconciler and
// attaches a default event handler to its Service informer.
func NewGenericServiceReconciler[T kmeta.OwnerRefable, L Lister[T]](ctx context.Context, gvk schema.GroupVersionKind,
	tracker tracker.Interface,
	adapterHandlerFn func(obj interface{}),
	ownersLister ListerGetter[T, L],
) GenericServiceReconciler[T, L] {

	serviceinformerv1.Get(ctx).Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGVK(gvk),
		Handler:    controller.HandleAll(adapterHandlerFn),
	})

	return newGenericServiceReconciler(ctx, tracker, ownersLister)
}

// NewMTGenericServiceReconciler creates a new GenericServiceReconciler for a
// multi-tenant adapter and attaches a default event handler to its Service
// informer.
func NewMTGenericServiceReconciler[T kmeta.OwnerRefable, L Lister[T]](ctx context.Context, typ kmeta.OwnerRefable,
	tracker tracker.Interface,
	adapterHandlerFn func(obj interface{}),
	ownersLister ListerGetter[T, L],
) GenericServiceReconciler[T, L] {

	serviceinformerv1.Get(ctx).Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: hasAdapterLabelsForType(typ),
		Handler:    controller.HandleAll(adapterHandlerFn),
	})

	return newGenericServiceReconciler(ctx, tracker, ownersLister)
}

// NewGenericRBACReconciler creates a new GenericRBACReconciler.
func NewGenericRBACReconciler[T kmeta.OwnerRefable, L Lister[T]](ctx context.Context,
	ownersLister ListerGetter[T, L],
) *GenericRBACReconciler[T, L] {

	return &GenericRBACReconciler[T, L]{
		SAClient:     k8sclient.Get(ctx).CoreV1().ServiceAccounts,
		RBClient:     k8sclient.Get(ctx).RbacV1().RoleBindings,
		SALister:     sainformerv1.Get(ctx).Lister().ServiceAccounts,
		RBLister:     rbinformerv1.Get(ctx).Lister().RoleBindings,
		OwnersLister: ownersLister,
	}
}

// newGenericServiceReconciler creates a new GenericServiceReconciler.
func newGenericServiceReconciler[T kmeta.OwnerRefable, L Lister[T]](ctx context.Context,
	tracker tracker.Interface,
	ownersLister ListerGetter[T, L],
) GenericServiceReconciler[T, L] {

	return GenericServiceReconciler[T, L]{
		SinkResolver:          resolver.NewURIResolverFromTracker(ctx, tracker),
		Client:                servingclient.Get(ctx).ServingV1().Services,
		Lister:                serviceinformerv1.Get(ctx).Lister().Services,
		GenericRBACReconciler: NewGenericRBACReconciler(ctx, ownersLister),
	}

}

// filteredGlobalResyncFunc is a function that enqueues all objects from the
// given informer that pass the provided filter function.
type filteredGlobalResyncFunc func(func(interface{}) bool, cache.SharedInformer)

// objectFilterFunc is a filtering function for filteredGlobalResyncFunc.
type objectFilterFunc func(interface{}) bool

// hasAdapterLabelsForType returns a function that filters based on standard
// labels applied to all adapters of the given component type.
func hasAdapterLabelsForType(typ kmeta.OwnerRefable) objectFilterFunc {
	return func(obj interface{}) bool {
		object, ok := obj.(metav1.Object)
		if !ok {
			return false
		}

		ls := CommonObjectLabels(typ).AsSelectorPreValidated()
		return ls.Matches(labels.Set(object.GetLabels()))
	}
}

// isInNamespace returns a filter function which returns whether the object
// passed to the filter is in the given namespace.
func isInNamespace(ns string) objectFilterFunc {
	return func(obj interface{}) bool {
		object := obj.(metav1.Object)
		return object.GetNamespace() == ns
	}
}

// EnqueueObjectsInNamespaceOf accepts an object and triggers a global resync
// of all objects in the given informer matching that object's namespace.
// Intended to be used to resync objects when the state of their (common)
// multi-tenant adapter changes.
func EnqueueObjectsInNamespaceOf(inf cache.SharedInformer, resyncFn filteredGlobalResyncFunc,
	logger *zap.SugaredLogger) func(interface{}) {

	return func(obj interface{}) {
		adapter, err := kmeta.DeletionHandlingAccessor(obj)
		if err != nil {
			logger.Error(err)
			return
		}

		resyncFn(isInNamespace(adapter.GetNamespace()), inf)
	}
}

// adapterPodWithAncestorOfKind returns a filter function which returns whether a Pod:
//   - has labels that correspond to an adapter of the given type
//   - has an outermost ancestor of the given kind
func adapterPodWithAncestorOfKind(ctx context.Context, typ kmeta.OwnerRefable) objectFilterFunc {
	return func(obj interface{}) bool {
		return hasAdapterLabelsForType(typ)(obj) &&
			hasOutermostAncestorOfKind(ctx, typ.GetGroupVersionKind())(obj)
	}
}

// hasOutermostAncestorOfKind returns a filter function which returns whether
// the given object has an outermost ancestor of the given kind.
func hasOutermostAncestorOfKind(ctx context.Context, gvk schema.GroupVersionKind) objectFilterFunc {
	return func(obj interface{}) bool {
		object, ok := obj.(metav1.Object)
		if !ok {
			return false
		}

		ancestorCtlrRef := outermostAncestorControllerRef(ctx, object)
		return ancestorCtlrRef != nil &&
			ancestorCtlrRef.APIVersion == gvk.GroupVersion().String() &&
			ancestorCtlrRef.Kind == gvk.Kind
	}
}

// outermostAncestorControllerRef returns the outermost ancestor controller of
// the given object.
func outermostAncestorControllerRef(ctx context.Context, obj metav1.Object) *metav1.OwnerReference {
	return resolveOutermostAncestorControllerRef(ctx, obj, schema.GroupVersionKind{})
}

// resolveOutermostAncestorControllerRef returns a reference to the controller
// of an API object, recursing up the hierarchy of controllers.
// eg. pod -> replicaset -> deployment -> TriggerMesh component
func resolveOutermostAncestorControllerRef(ctx context.Context, obj metav1.Object, gvk schema.GroupVersionKind) *metav1.OwnerReference {
	controllerRef := metav1.GetControllerOf(obj)
	if controllerRef == nil {
		if firstIteration := gvk.Empty(); firstIteration {
			return nil
		}
		self := metav1.NewControllerRef(obj, gvk)
		return self
	}

	var controllerObj metav1.Object
	var err error

	switch controllerRef.Kind {
	case "ReplicaSet":
		controllerObj, err = replicasetinformerv1.Get(ctx).Lister().ReplicaSets(obj.GetNamespace()).Get(controllerRef.Name)
	case "Deployment":
		controllerObj, err = deploymentinformerv1.Get(ctx).Lister().Deployments(obj.GetNamespace()).Get(controllerRef.Name)
	default:
		// we only support a subset of dependency chains
		return controllerRef
	}
	if err != nil {
		return nil
	}

	gvk = schema.FromAPIVersionAndKind(controllerRef.APIVersion, controllerRef.Kind)
	return resolveOutermostAncestorControllerRef(ctx, controllerObj, gvk)
}

// podOutermostAncestorHandlerFn returns a resource handler function which passes
// the outermost ancestor of a Pod to the provided resource handler function.
func podOutermostAncestorHandlerFn(ctx context.Context, handlerFn func(interface{})) func(interface{}) {
	return func(obj interface{}) {
		object, ok := obj.(metav1.Object)
		if !ok {
			return
		}

		ancestorCtlrRef := outermostAncestorControllerRef(ctx, object)
		if ancestorCtlrRef == nil {
			return
		}

		// It is assumed that handlerFn is impl.EnqueueControllerOf,
		// originally passed by the component's Reconciler implementation,
		// which means that the type of the variable passed to it must
		//  - satisfy kmeta.Accessor
		//  - be controlled by the object to be enqueued
		handlerFn(newAccessorWithController(object.GetNamespace(), ancestorCtlrRef))
	}
}

// newAccessorWithController returns a kmeta.Accessor that has the given owner.
func newAccessorWithController(ns string, owner *metav1.OwnerReference) kmeta.Accessor {
	return &accessor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ns,
			OwnerReferences: []metav1.OwnerReference{*owner},
		},
	}
}

// accessor is an adapter type which allows a metav1.Object to implement kmeta.Accessor.
type accessor struct {
	metav1.ObjectMeta
}

var _ kmeta.Accessor = (*accessor)(nil)

// GroupVersionKind implements kmeta.Accessor.
func (*accessor) GroupVersionKind() schema.GroupVersionKind { return schema.GroupVersionKind{} }

// SetGroupVersionKind implements kmeta.Accessor.
func (*accessor) SetGroupVersionKind(schema.GroupVersionKind) {}

// GetObjectKind implements kmeta.Accessor.
func (*accessor) GetObjectKind() schema.ObjectKind { return nil }

// DeepCopyObject implements kmeta.Accessor.
func (*accessor) DeepCopyObject() runtime.Object { return nil }
