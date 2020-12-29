package instance_controller

import (
	"context"
	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	virtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("LabOperator controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		InstanceName      = "test-instance"
		InstanceNamespace = "instance-namespace"
		TemplateName      = "test-template"
		TemplateNamespace = "template-namespace"
		TenantName        = "test-tenant"

		timeout  = time.Second * 20
		interval = time.Millisecond * 500
	)

	var (
		ctx        context.Context
		templateNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: TemplateNamespace,
				Labels: map[string]string{
					"test-suite": "true",
				},
			},
			Spec:   v1.NamespaceSpec{},
			Status: v1.NamespaceStatus{},
		}
		instanceNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: InstanceNamespace,
				Labels: map[string]string{
					"production": "true",
					"test-suite": "true",
				},
			},
			Spec:   v1.NamespaceSpec{},
			Status: v1.NamespaceStatus{},
		}
		webdavSecret = v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webdav-secret",
				Namespace: InstanceNamespace,
			},
			Data: map[string][]byte{
				"username": []byte("username"),
				"password": []byte("password"),
			},
			StringData: nil,
			Type:       "",
		}
		instance = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Template: crownlabsv1alpha2.GenericRef{
					Name:      TemplateName,
					Namespace: TemplateNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name: TenantName,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}
		templateSpecSingleVM = crownlabsv1alpha2.TemplateSpec{
			WorkspaceRef: crownlabsv1alpha2.GenericRef{},
			PrettyName:   "Wonderful Template",
			Description:  "A description",
			EnvironmentList: []crownlabsv1alpha2.Environment{
				{
					Name:       "Test",
					GuiEnabled: true,
					Resources: crownlabsv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: crownlabsv1alpha2.ClassVM,
					Persistent:      false,
					Image:           "trololo/vm",
				},
			},
			DeleteAfter: "",
		}
		tenant = crownlabsv1alpha1.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name: TenantName,
			},
			Spec: crownlabsv1alpha1.TenantSpec{
				FirstName:  "Mario",
				LastName:   "Rossi",
				Email:      "mario@rossi.com",
				Workspaces: nil,
				PublicKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDP5O4CX17GK17GN+xxfoUpjz6s1FLLQdVSBYSJ02uS/HHueJmE8TQS8tpNfVHQ+i9cCpR+RMXUUDscNjgTxcF8Z0iRIfX6InRUbJt7FSYERX3roSy4YyhNnkDhQe9+1cQNhZsUtVKkTAE4Ew/dnqHjAKMjz+nhiNio7bL0ZsODGZ90uFUfzcg2RUyftluN+IaX9cD9VvmXRmKFMiIUkebOnLxREiOS6aqe8NrqbkK6Bkt0hWhr8U2pfDK86BbgFGL9e7Ms1dWxDPfrOVLnteN0Xe0bLv3HoW1KVlnoC4hwSHaRlhlSU0wgTkfvPzQy/eM95oTrQQCr0fmvjv5uiciP",
				},
				CreateSandbox: false,
			},
			Status: crownlabsv1alpha1.TenantStatus{},
		}
		ns1  = v1.Namespace{}
		tmp1 = crownlabsv1alpha2.Template{}
	)

	Context("", func() {
		It("Setting up the Instance and Template namespaces", func() {
			ctx = context.Background()
			Expect(k8sClient.Create(ctx, &templateNs)).Should(Succeed())
			doesEventuallyExists(ctx, types.NamespacedName{
				Name: TemplateNamespace,
			}, &ns1, BeTrue(), timeout, interval)
			Expect(k8sClient.Create(ctx, &instanceNs)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &webdavSecret)).Should(Succeed())
			doesEventuallyExists(ctx, types.NamespacedName{
				Name: InstanceNamespace,
			}, &ns1, BeTrue(), timeout, interval)
			Expect(k8sClient.Create(ctx, &tenant)).Should(Succeed())
			doesEventuallyExists(ctx, types.NamespacedName{
				Name: TenantName,
			}, &tenant, BeTrue(), timeout, interval)
		})
		Context("Create an Instance from a single VM template", func() {
			It("Creates the single-vm Template and the instance", func() {
				By("Creating the Template")
				template := crownlabsv1alpha2.Template{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      TemplateName,
						Namespace: TemplateNamespace,
					},
					Spec:   templateSpecSingleVM,
					Status: crownlabsv1alpha2.TemplateStatus{},
				}
				Expect(k8sClient.Create(ctx, &template)).Should(Succeed())
				doesEventuallyExists(ctx, types.NamespacedName{
					Name:      TemplateName,
					Namespace: TemplateNamespace,
				}, &tmp1, BeTrue(), timeout, interval)

				By("and creating a Instance associated to the template")

				Expect(k8sClient.Create(ctx, &instance)).Should(Succeed())
				var vmList virtv1.VirtualMachineInstanceList
				m := map[string]string{
					"instance-name": InstanceName,
				}

				By("VirtualMachine Should Exists")
				Eventually(func() bool {
					err := k8sClient.List(ctx, &vmList, client.MatchingLabels(m))
					if err != nil {
						Fail("Unable to list VirtualMachines")
					}
					return len(vmList.Items) == 1
				}, timeout, interval).Should(BeTrue())
				By("VirtualMachine Has An Ownerreferce")

				flag := true
				expectedOwnerReference := metav1.OwnerReference{
					Kind:               "Instance",
					APIVersion:         "crownlabs.polito.it/v1alpha2",
					UID:                instance.ObjectMeta.UID,
					Name:               InstanceName,
					Controller:         &flag,
					BlockOwnerDeletion: &flag,
				}
				Expect(vmList.Items[0].ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

				By("Cleaning Instance")
				Expect(k8sClient.Delete(ctx, &instance)).Should(Succeed())
				Eventually(func() bool {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Namespace: InstanceName,
						Name:      InstanceName,
					}, &instance)
					if err != nil && errors2.IsNotFound(err) {
						return true
					}
					return false
				}, timeout, interval).Should(BeTrue())
			})
		})
	})
})

func doesEventuallyExists(ctx context.Context, nsLookupKey types.NamespacedName, createdObject runtime.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout time.Duration, interval time.Duration) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, nsLookupKey, createdObject)
		return err == nil
	}, timeout, interval).Should(expectedStatus)

}