// +build !ignore_autogenerated

/*
Copyright 2019 The Knative Authors

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

// Code generated by reconciler-gen. DO NOT EDIT.

package addressableservice

import (
	"context"
	"encoding/json"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/controller"
	"reflect"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
	"knative.dev/sample-controller/pkg/apis/samples/v1alpha1"
	clientset "knative.dev/sample-controller/pkg/client/clientset/versioned"
	listers "knative.dev/sample-controller/pkg/client/listers/samples/v1alpha1"
)

// Interface defines the strongly typed interfaces to be implemented by a
// controller reconciling the Kind.
type Interface interface {
	// ReconcileKind implements custom logic to reconcile the Kind. Any changes
	// to the objects .Status or .Finalizers will be propaged to the stored
	// object. It is recommended that implementors do not call any update calls
	// for the Kind inside of ReconcileKind, it is the resonsbility of the core
	// controller to propagate those properties.
	ReconcileKind(ctx context.Context, asvc *v1alpha1.AddressableService) error
}

// Reconciler implements controller.Reconciler for AddressableService resources.
type Core struct {
	// Client is used to write back status updates.
	Client clientset.Interface

	// Listers index properties about resources
	Lister listers.AddressableServiceLister

	// The tracker builds an index of what resources are watching other
	// resources so that we can immediately react to changes to changes in
	// tracked resources.
	Tracker tracker.Interface

	// Recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	Recorder record.EventRecorder

	// Reconciler is the implementation of the business logic of the resource.
	Reconciler Interface

	// FinalizerName is the name of the finalizer to use when finalizing the
	// resource.
	FinalizerName string
}

// Check that our Core implements controller.Reconciler
var _ controller.Reconciler = (*Core)(nil)

// Reconcile implements controller.Reconciler
func (r *Core) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Errorf("invalid resource key: %s", key)
		return nil
	}

	// If our controller has configuration state, we'd "freeze" it and
	// attach the frozen configuration to the context.
	//    ctx = r.configStore.ToContext(ctx)

	// Get the resource with this namespace/name.
	original, err := r.Lister.AddressableServices(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		// The resource may no longer exist, in which case we stop processing.
		logger.Errorf("resource %q no longer exists", key)
		return nil
	} else if err != nil {
		return err
	}
	// Don't modify the informers copy.
	resource := original.DeepCopy()

	// Reconcile this copy of the resource and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.Reconciler.ReconcileKind(ctx, resource)

	// Syncronize the finalizers.
	if equality.Semantic.DeepEqual(original.Finalizers, resource.Finalizers) {
		// If we didn't change finalizers then don't call updateFinalizers.
	} else if _, updated, fErr := r.updateFinalizers(ctx, resource); fErr != nil {
		logger.Warnw("Failed to update finalizers", zap.Error(fErr))
		r.Recorder.Eventf(resource, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update finalizers for %q: %v", resource.Name, fErr)
		return fErr
	} else if updated {
		// There was a difference and updateFinalizers said it updated and did not return an error.
		r.Recorder.Eventf(resource, corev1.EventTypeNormal, "Updated", "Updated %q finalizers", resource.GetName())
	}

	// Syncronize the status.
	if equality.Semantic.DeepEqual(original.Status, resource.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err = r.updateStatus(resource); err != nil {
		logger.Warnw("Failed to update resource status", zap.Error(err))
		r.Recorder.Eventf(resource, corev1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for %q: %v", resource.Name, err)
		return err
	}

	// Report the reconciler error, if any.
	if reconcileErr != nil {
		r.Recorder.Event(resource, corev1.EventTypeWarning, "InternalError", reconcileErr.Error())
	}
	return reconcileErr
}

// Update the Status of the resource.  Caller is responsible for checking
// for semantic differences before calling.
func (r *Core) updateStatus(desired *v1alpha1.AddressableService) (*v1alpha1.AddressableService, error) {
	actual, err := r.Lister.AddressableServices(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(actual.Status, desired.Status) {
		return actual, nil
	}
	// Don't modify the informers copy
	existing := actual.DeepCopy()
	existing.Status = desired.Status
	return r.Client.SamplesV1alpha1().AddressableServices(desired.Namespace).UpdateStatus(existing)
}

// Update the Finalizers of the resource.
func (r *Core) updateFinalizers(ctx context.Context, desired *v1alpha1.AddressableService) (*v1alpha1.AddressableService, bool, error) {
	actual, err := r.Lister.AddressableServices(desired.Namespace).Get(desired.Name)
	if err != nil {
		return nil, false, err
	}

	// Don't modify the informers copy.
	existing := actual.DeepCopy()

	var finalizers []string

	// If there's nothing to update, just return.
	existingFinalizers := sets.NewString(existing.Finalizers...)
	desiredFinalizers := sets.NewString(desired.Finalizers...)

	if desiredFinalizers.Has(r.FinalizerName) {
		if existingFinalizers.Has(r.FinalizerName) {
			// Nothing to do.
			return desired, false, nil
		}
		// Add the finalizer.
		finalizers = append(existing.Finalizers, r.FinalizerName)
	} else {
		if !existingFinalizers.Has(r.FinalizerName) {
			// Nothing to do.
			return desired, false, nil
		}
		// Remove the finalizer.
		existingFinalizers.Delete(r.FinalizerName)
		finalizers = existingFinalizers.List()
	}

	mergePatch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"finalizers":      finalizers,
			"resourceVersion": existing.ResourceVersion,
		},
	}

	patch, err := json.Marshal(mergePatch)
	if err != nil {
		return desired, false, err
	}

	update, err := r.Client.SamplesV1alpha1().AddressableServices(desired.Namespace).Patch(existing.Name, types.MergePatchType, patch)
	return update, true, err
}

func (r *Core) setFinalizer(a *v1alpha1.AddressableService) {
	finalizers := sets.NewString(a.Finalizers...)
	finalizers.Insert(r.FinalizerName)
	a.Finalizers = finalizers.List()
}

func (r *Core) unsetFinalizer(a *v1alpha1.AddressableService) {
	finalizers := sets.NewString(a.Finalizers...)
	finalizers.Delete(r.FinalizerName)
	a.Finalizers = finalizers.List()
}
