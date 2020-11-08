/*


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
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tenantv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/tenant-operator/api/v1alpha1"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=workspaces/status,verbs=get;update;patch

func (r *WorkspaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("workspace", req.NamespacedName)

	var ws tenantv1alpha1.Workspace

	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// reconcile was triggered by a delete request
		log.Info(fmt.Sprintf("Workspace %s deleted", req.Name))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	wsnsName := fmt.Sprintf("workspace-%s", req.Name)
	namespaceName := types.NamespacedName{
		Name:      wsnsName,
		Namespace: "",
	}

	wsns := v1.Namespace{}
	wsns.Name = wsnsName
	wsns.Labels = make(map[string]string)
	pTrue := true
	wsOwnerRef :=
		[]metav1.OwnerReference{
			{
				APIVersion:         ws.APIVersion,
				Kind:               ws.Kind,
				Name:               ws.Name,
				UID:                ws.UID,
				BlockOwnerDeletion: &pTrue,
			}}
	wsns.OwnerReferences = append(wsns.OwnerReferences, wsOwnerRef...)
	wsns.Labels["type"] = "workspace"

	if err := r.Get(ctx, namespaceName, &wsns); err != nil {
		// namespace was not defined, need to create it
		if err := r.Create(ctx, &wsns); err != nil {
			log.Error(err, "Unable to create namespace")
			ws.Status.Namespace.Created = false
			ws.Status.Namespace.Name = ""
			if err := r.Status().Update(ctx, &ws); err != nil {
				log.Error(err, "Unable to update status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("Namespace %s created", wsnsName))

	} else {
		if err := r.Update(ctx, &wsns); err != nil {
			log.Error(err, "Unable to update namespace")
			ws.Status.Namespace.Created = false
			ws.Status.Namespace.Name = ""
			if err := r.Status().Update(ctx, &ws); err != nil {
				log.Error(err, "Unable to update status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		log.Info(fmt.Sprintf("Namespace %s updated", wsnsName))
	}

	// update status of workspace with info about namespace
	ws.Status.Namespace.Created = true
	ws.Status.Namespace.Name = wsnsName
	if err := r.Status().Update(ctx, &ws); err != nil {
		log.Error(err, "Unable to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1alpha1.Workspace{}).
		Complete(r)
}

func updateStatusNamespace(r *WorkspaceReconciler, ctx context.Context, name string) {
}
