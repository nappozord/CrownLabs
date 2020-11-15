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
	"strings"

	gocloak "github.com/Nerzal/gocloak/v7"
	"github.com/go-logr/logr"
	tenantv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/tenant-operator/api/v1alpha1"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	KcClient gocloak.GoCloak
	KcToken  *gocloak.JWT
}

// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=workspaces/status,verbs=get;update;patch

func (r *WorkspaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("workspace", req.NamespacedName)

	kcToken := r.KcToken.AccessToken
	targetClient := "k8s"
	var targetClientID string

	clients, err := r.KcClient.GetClients(ctx, kcToken, "crownlabs", gocloak.GetClientsParams{ClientID: &targetClient})
	if err != nil {
		log.Error(err, "Error when getting k8s client")
	} else if len(clients) > 1 {
		log.Error(nil, "too many k8s clients")

	} else if len(clients) < 0 {
		log.Error(nil, "no k8s client")

	} else {
		targetClientID = *clients[0].ID
		log.Info("Got client id", "id", targetClientID)
	}

	var ws tenantv1alpha1.Workspace

	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// reconcile was triggered by a delete request
		log.Info(fmt.Sprintf("Workspace %s deleted", req.Name))
		deleteWorkspaceRoles(ctx, r.KcClient, kcToken, targetClientID, req.Name)
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
		log.Error(err, "Unable to create or update namespace")
		// update status of workspace with info about namespace. failed
		ws.Status.Namespace.Created = false
		ws.Status.Namespace.Name = ""
		if err := r.Status().Update(ctx, &ws); err != nil {
			log.Error(err, "Unable to update status")
			return ctrl.Result{}, err
		}
	}
	// update status of workspace with info about namespace, success
	ws.Status.Namespace.Created = true
	ws.Status.Namespace.Name = nsName
	if err := r.Status().Update(ctx, &ws); err != nil {
		log.Error(err, "Unable to update status")
		return ctrl.Result{}, err
	}

	err = createKcRoleForWorkspace(ctx, r.KcClient, kcToken, targetClientID, ws.Name)
	if err != nil {
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

func createKcRoleForWorkspace(ctx context.Context, kcClient gocloak.GoCloak, token string, targetClientID string, wsName string) error {
	newUserRoleName := fmt.Sprintf("workspace-%s:user", wsName)
	err := createKcRole(ctx, kcClient, token, targetClientID, newUserRoleName)
	if err != nil {
		log.Error("Could not create user role")
		return err
	}
	newAdminRoleName := fmt.Sprintf("workspace-%s:admin", wsName)
	err = createKcRole(ctx, kcClient, token, targetClientID, newAdminRoleName)
	if err != nil {
		log.Error("Could not create admin role")
		return err
	}
	return nil
}

func createKcRole(ctx context.Context, kcClient gocloak.GoCloak, token string, targetClientID string, newRoleName string) error {
	// check if keycloak role already esists

	_, err := kcClient.GetClientRole(ctx, token, "crownlabs", targetClientID, newRoleName)
	if err != nil && strings.Contains(err.Error(), "404 Not Found: Could not find role") {
		// error corresponds to "not found"
		// need to create new role
		log.Info("Role didn't exists", "role", newRoleName)
		tr := true
		createdRoleName, err := kcClient.CreateClientRole(ctx, token, "crownlabs", targetClientID, gocloak.Role{Name: &newRoleName, ClientRole: &tr})
		if err != nil {
			log.Error(err, "Error when creating role")
			return err
		}
		log.Info("Role created", "rolename", createdRoleName)
		return nil
	} else if err != nil {
		log.Error(err, "Error when getting user role")
		return err
	} else {
		log.Info("Role already existed", "role", newRoleName)
		return nil
	}
}

func deleteWorkspaceRoles(ctx context.Context, kcClient gocloak.GoCloak, token string, targetClientID string, wsName string) error {
	userRoleName := fmt.Sprintf("workspace-%s:user", wsName)
	err := kcClient.DeleteClientRole(ctx, token, "crownlabs", targetClientID, userRoleName)
	if err != nil {
		log.Error("Could not delete user role")
		return err
	}
	adminRoleName := fmt.Sprintf("workspace-%s:admin", wsName)
	err = kcClient.DeleteClientRole(ctx, token, "crownlabs", targetClientID, adminRoleName)
	if err != nil {
		log.Error("Could not delete admin role")
		return err
	}
	return nil
}
