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

	gocloak "github.com/Nerzal/gocloak/v7"
	tenantv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/tenant-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	KcClient gocloak.GoCloak
	KcToken  *gocloak.JWT
}

// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=workspaces/status,verbs=get;update;patch

// Reconcile reconciles the state of a workspace resource
func (r *WorkspaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	klog.Info("HELLOOOO")

	CheckAndRenewToken(ctx, r.KcClient, &r.KcToken)
	token := r.KcToken.AccessToken

	targetClientID, err := GetClientID(ctx, r.KcClient, token, "crownlabs", "k8s")
	if err != nil {
		klog.Error(err, "Error when getting client")
		return ctrl.Result{}, err
	}

	var ws tenantv1alpha1.Workspace

	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// reconcile was triggered by a delete request
		klog.Info(fmt.Sprintf("Workspace %s deleted", req.Name))
		if err := deleteWorkspaceRoles(ctx, r.KcClient, token, targetClientID, req.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nsName := fmt.Sprintf("workspace-%s", ws.Name)
	ns := v1.Namespace{}
	ns.Name = nsName
	ns.Namespace = ""
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		updateNamespace(ws, &ns, nsName)
		return ctrl.SetControllerReference(&ws, &ns, r.Scheme)
	}); err != nil {
		klog.Error(err, "Unable to create or update namespace")
		// update status of workspace with info about namespace. failed
		ws.Status.Namespace.Created = false
		ws.Status.Namespace.Name = ""
		if err := r.Status().Update(ctx, &ws); err != nil {
			klog.Error(err, "Unable to update status")
			return ctrl.Result{}, err
		}
	}
	// update status of workspace with info about namespace, success
	ws.Status.Namespace.Created = true
	ws.Status.Namespace.Name = nsName
	if err := r.Status().Update(ctx, &ws); err != nil {
		klog.Error(err, "Unable to update status")
		return ctrl.Result{}, err
	}

	if err := createKcRolesForWorkspace(ctx, r.KcClient, token, "crownlabs", targetClientID, ws.Name); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1alpha1.Workspace{}).
		Complete(r)
}

func updateNamespace(ws tenantv1alpha1.Workspace, ns *v1.Namespace, wsnsName string) {
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}
	ns.Labels["type"] = "workspace"
}

func createKcRolesForWorkspace(ctx context.Context, kcClient gocloak.GoCloak, token string, realmName string, targetClientID string, wsName string) error {
	newUserRoleName := fmt.Sprintf("workspace-%s:user", wsName)

	if err := createKcRole(ctx, kcClient, token, realmName, targetClientID, newUserRoleName); err != nil {
		klog.Error("Could not create user role")
		return err
	}
	newAdminRoleName := fmt.Sprintf("workspace-%s:admin", wsName)

	if err := createKcRole(ctx, kcClient, token, realmName, targetClientID, newAdminRoleName); err != nil {
		klog.Error("Could not create admin role")
		return err
	}
	return nil
}

func deleteWorkspaceRoles(ctx context.Context, kcClient gocloak.GoCloak, token string, targetClientID string, wsName string) error {

	userRoleName := fmt.Sprintf("workspace-%s:user", wsName)
	if err := kcClient.DeleteClientRole(ctx, token, "crownlabs", targetClientID, userRoleName); err != nil {
		klog.Error("Could not delete user role")
		return err
	}

	adminRoleName := fmt.Sprintf("workspace-%s:admin", wsName)
	if err := kcClient.DeleteClientRole(ctx, token, "crownlabs", targetClientID, adminRoleName); err != nil {
		klog.Error("Could not delete admin role")
		return err
	}
	return nil
}
