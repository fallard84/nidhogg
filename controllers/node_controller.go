/*
Copyright 2022.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"uswitch.com/nidhogg/pkg/nidhogg"
)

// type nodeEnqueue struct{}

// // Update implements the interface
// func (e *nodeEnqueue) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {}

// // Delete implements the interface
// func (e *nodeEnqueue) Delete(_ event.DeleteEvent, _ workqueue.RateLimitingInterface) {}

// // Generic implements the interface
// func (e *nodeEnqueue) Generic(_ event.GenericEvent, _ workqueue.RateLimitingInterface) {}

// // Create adds the node to the queue, the node is created as NotReady and without daemonset pods
// func (e *nodeEnqueue) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
// 	if evt.Object == nil {
// 		return
// 	}
// 	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
// 		Name: evt.Object.GetName(),
// 	}})
// }

type podEnqueue struct{}

// canAddToQueue check if the Pod is associated to a node and is a daemonset pod
func (e *podEnqueue) canAddToQueue(pod *corev1.Pod) bool {
	log.Log.Info("Checkinf if pod should be queued", "pod", pod)
	if pod.Spec.NodeName == "" {
		return false
	}
	owner := v1.GetControllerOf(pod)
	if owner == nil {
		return false
	}
	log.Log.Info("Checkinf kind", "kind", owner.Kind)
	return owner.Kind == "DaemonSet"
}

// Generic implements the interface
func (e *podEnqueue) Generic(_ event.GenericEvent, _ workqueue.RateLimitingInterface) {}

// Create adds the node of the daemonset pod to the queue
func (e *podEnqueue) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	pod, ok := evt.Object.(*corev1.Pod)
	if !ok {
		return
	}
	if !e.canAddToQueue(pod) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: pod.Spec.NodeName,
	}})

}

// Update adds the node of the updated daemonset pod to the queue
func (e *podEnqueue) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	pod, ok := evt.ObjectNew.(*corev1.Pod)
	if !ok {
		return
	}
	if !e.canAddToQueue(pod) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: pod.Spec.NodeName,
	}})
}

// Delete adds the node of the deleted daemonset pod to the queue
func (e *podEnqueue) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	pod, ok := evt.Object.(*corev1.Pod)
	if !ok {
		return
	}
	if !e.canAddToQueue(pod) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: pod.Spec.NodeName,
	}})
}

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	Handler *nidhogg.Handler
	Scheme  *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Node object and makes changes based on the state read
// and what is in the Node.Spec
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=nodes/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Node object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Node instance
	instance := &corev1.Node{}
	logger.Info("Reconciling", "Instance", instance)
	err := r.Handler.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	logger.Info("Handling Node", "Instance", instance)
	return r.Handler.HandleNode(instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&corev1.Node{}).
		// Watches(&source.Kind{Type: &corev1.Node{}}, &nodeEnqueue{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, &podEnqueue{}).
		Complete(r)
}
