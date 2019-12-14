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

// Code generated by main. DO NOT EDIT.

package addressableservice

import (
	"context"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	informer "knative.dev/pkg/apis/duck/informer"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"
)

const (
	controllerAgentName = "unnameable_-controller"
	finalizerName       = "unnameable_"
)

func NewImpl(ctx context.Context, r *Reconciler) *controller.Impl {
	logger := logging.FromContext(ctx)

	impl := controller.NewImpl(r, logger, "unnameable_")

	informer := informer.Get(ctx)

	r.Core = Core{
		Client:  asclient.Get(ctx),
		Lister:  informer.Lister(),
		Tracker: tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx)),
		Recorder: record.NewBroadcaster().NewRecorder(
			scheme.Scheme, corev1.EventSource{Component: controllerAgentName}),
		FinalizerName: finalizerName,
		Reconciler:    r,
	}

	logger.Info("Setting up core event handlers")
	asInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	return impl
}
