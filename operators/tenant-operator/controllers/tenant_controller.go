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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tenantv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/tenant-operator/api/v1alpha1"
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.crownlabs.polito.it,resources=tenants/status,verbs=get;update;patch

func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("tenant", req.NamespacedName)

	log.Info(fmt.Sprintf("%s has been created", req.Name))
	var tenant tenantv1alpha1.Tenant

	r.Get(ctx, req.NamespacedName, &tenant)
	log.Info(fmt.Sprintf("Reached %v", tenant.Status))
	if tenant.Status.Subscriptions == nil {
		tenant.Status.Subscriptions = make(map[string]tenantv1alpha1.SubscriptionStatus)
	}
	tenant.Status.Subscriptions["keycloak"] = tenantv1alpha1.Pending
	log.Info(fmt.Sprintf("New status %v", tenant.Status))
	if err := r.Status().Update(ctx, &tenant); err != nil {
		log.Error(err, "failed to update status")
	} else {
		log.Info("UPDATED STATUS")
	}
	return ctrl.Result{}, nil
}

func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1alpha1.Tenant{}).
		Complete(r)
}
