// SPDX-License-Identifier: Apache-2.0
// Copyright 2025 BoanLab @ DKU

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	securityv1 "kubefort-operator/api/v1"
)

var _ = Describe("KubeFortPolicy Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		kubefortpolicy := &securityv1.KubeFortPolicy{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind KubeFortPolicy")
			err := k8sClient.Get(ctx, typeNamespacedName, kubefortpolicy)
			if err != nil && errors.IsNotFound(err) {
				resource := &securityv1.KubeFortPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: securityv1.KubeFortPolicySpec{
						Action: "Allow",
						Selector: map[string]string{
							"app": "test",
						},
						Process: []securityv1.ProcessRule{
							{
								Path: "/bin/test",
							},
						},
						File: []securityv1.FileRule{
							{
								Path:     "/etc/test",
								ReadOnly: true,
							},
						},
						Network: []securityv1.NetworkRule{
							{
								Direction: "ingress",
								IPBlock: securityv1.IPBlock{
									CIDR: "10.0.0.0/24",
								},
								Ports: []securityv1.Port{
									{
										Protocol: "TCP",
										Port:     80,
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &securityv1.KubeFortPolicy{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err != nil {
				if errors.IsNotFound(err) {
					// Resource is already deleted, which is fine
					return
				}
				Expect(err).NotTo(HaveOccurred())
			}

			By("Cleanup the specific resource instance KubeFortPolicy")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &KubeFortPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify the policy status
			Eventually(func() string {
				var policy securityv1.KubeFortPolicy
				err := k8sClient.Get(ctx, typeNamespacedName, &policy)
				if err != nil {
					return ""
				}
				return policy.Status.PolicyStatus
			}, time.Second*10, time.Second).Should(Equal("Active"))
		})

		It("should handle policy updates", func() {
			By("Updating the policy")
			var policy securityv1.KubeFortPolicy
			err := k8sClient.Get(ctx, typeNamespacedName, &policy)
			Expect(err).NotTo(HaveOccurred())

			policy.Spec.Action = "Audit"
			Expect(k8sClient.Update(ctx, &policy)).To(Succeed())

			By("Reconciling the updated resource")
			controllerReconciler := &KubeFortPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify the policy status after update
			Eventually(func() string {
				var updatedPolicy securityv1.KubeFortPolicy
				err := k8sClient.Get(ctx, typeNamespacedName, &updatedPolicy)
				if err != nil {
					return ""
				}
				return updatedPolicy.Status.PolicyStatus
			}, time.Second*10, time.Second).Should(Equal("Active"))
		})

		It("should handle policy deletion", func() {
			By("Deleting the policy")
			var policy securityv1.KubeFortPolicy
			err := k8sClient.Get(ctx, typeNamespacedName, &policy)
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Delete(ctx, &policy)).To(Succeed())

			By("Reconciling the deleted resource")
			controllerReconciler := &KubeFortPolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify the policy is deleted
			Eventually(func() bool {
				var deletedPolicy securityv1.KubeFortPolicy
				err := k8sClient.Get(ctx, typeNamespacedName, &deletedPolicy)
				return errors.IsNotFound(err)
			}, time.Second*10, time.Second).Should(BeTrue())
		})

		It("should handle invalid policy actions", func() {
			By("Creating a policy with invalid action")
			invalidPolicy := &securityv1.KubeFortPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-policy",
					Namespace: "default",
				},
				Spec: securityv1.KubeFortPolicySpec{
					Action: "InvalidAction",
				},
			}
			err := k8sClient.Create(ctx, invalidPolicy)
			Expect(err).To(HaveOccurred())
			Expect(errors.IsInvalid(err)).To(BeTrue())
		})
	})
})
