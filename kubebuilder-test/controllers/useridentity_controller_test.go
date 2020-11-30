package controllers

import (
	"context"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GroupRoleBinding", func() {

	var reconciler UserIdentityReconciler

	var key types.NamespacedName
	var req types.NamespacedName

	BeforeEach(func() {
		k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).ToNot(HaveOccurred())
		k8sClient = k8sManager.GetClient()

		reconciler = UserIdentityReconciler{
				Client: k8sClient,
				Log:    ctrl.Log.WithName("controllers").WithName("UserIdentity"),
				Scheme: k8sManager.GetScheme(),
		}

		Expect(reconciler.SetupWithManager(k8sManager)).NotTo(HaveOccurred(), "failed to set up the manager")

	})

	When("A UserIdentity is created", func() {

		It("Should find a GroupRoleBinding", func() {
			Eventually(func() bool {
				fetched := &rbacv1.ClusterRoleBinding{}
				_ = k8sClient.Get(context.TODO(), req, fetched)
				return fetched != nil
			}, 10, 1).Should(Succeed())
		})
	})
})
