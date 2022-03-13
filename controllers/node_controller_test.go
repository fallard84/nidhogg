package controllers

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
var _ = Describe("NodeController", func() {
	const (
		taintKey = "nidhogg.uswitch.com"
		timeout  = time.Second * 30
		duration = time.Second * 30
		interval = time.Millisecond * 250
	)
	Context("When mandatory daemonset is not running", func() {

		It("Nodes should get a taint", func() {

			ctx := context.Background()
			instance := &corev1.Node{}
			namespacedName := types.NamespacedName{
				Namespace: "",
				Name:      "",
			}
			// Check that the claim was created succesfully
			Eventually(func() bool {
				k8sClient.Get(ctx, namespacedName, instance)
				log.Log.Info("checking instance", "instance", instance)
				for _, taint := range instance.Spec.Taints {
					if strings.HasPrefix(taint.Key, taintKey) {
						return true
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())

			// Check that the operator created a secret for the claim
			// Eventually(func() bool {
			// 	err := k8sClient.Get(ctx, fooSecretLookupKey, retrievedSecret)
			// 	return err == nil
			// }, timeout, interval).Should(BeTrue())
			// json.Unmarshal(retrievedSecret.Data["password.json"], &secret)
			// Expect(secret.Infra.Db.User).Should(Equal("proxyuser-default"))
			// Expect(secret.Infra.Db.Password).Should(Equal("zaskwef9139kd"))

			// By("Deleting the secret owned by the claim")
			// Expect(k8sClient.Delete(ctx, retrievedSecret)).Should(Succeed())
			// // Check that the secret gets recreated
			// Eventually(func() bool {
			// 	err := k8sClient.Get(ctx, fooSecretLookupKey, retrievedSecret)
			// 	return err == nil
			// }, timeout, interval).Should(BeTrue())
			// json.Unmarshal(retrievedSecret.Data["password.json"], &secret)
			// Expect(secret.Infra.Db.Password).Should(Equal("zaskwef9139kd"))

			// By("Deleting a claim")
			// Expect(k8sClient.Delete(ctx, tenantDatabaseUserClaim)).Should(Succeed())
			// // Check that the claim is gone
			// Eventually(func() bool {
			// 	err := k8sClient.Get(ctx, claimLookupKey, retrievedClaim)
			// 	return err != nil
			// }, timeout, interval).Should(BeTrue())
			// // Check that no secret exists
			// Eventually(func() bool {
			// 	err := k8sClient.Get(ctx, fooSecretLookupKey, retrievedSecret)
			// 	return err != nil
			// }, timeout, interval).Should(BeTrue())
		})
	})
})
