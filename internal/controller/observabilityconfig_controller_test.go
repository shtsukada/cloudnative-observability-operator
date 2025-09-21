/*
Copyright 2025 shtsukada.

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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	observabilityv1alpha1 "github.com/shtsukada/cloudnative-observability-operator/api/v1alpha1"
	conditions "github.com/shtsukada/cloudnative-observability-operator/internal/shared/conditions"
)

var _ = Describe("ObservabilityConfig Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const namespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace, // TODO(user):Modify as needed
		}
		observabilityconfig := &observabilityv1alpha1.ObservabilityConfig{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ObservabilityConfig")
			err := k8sClient.Get(ctx, typeNamespacedName, observabilityconfig)
			if err != nil && errors.IsNotFound(err) {
				resource := &observabilityv1alpha1.ObservabilityConfig{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: observabilityv1alpha1.ObservabilityConfigSpec{
						Endpoint: "otel-collector.monitoring:4317",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("cleanup the specific resource instance observabilityConfig")
			resource := &observabilityv1alpha1.ObservabilityConfig{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should set Conditions/Phase/ObservedGeneration as expected", func() {
			r := &ObservabilityConfigReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: record.NewFakeRecorder(32),
			}

			_, err := r.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			oc := &observabilityv1alpha1.ObservabilityConfig{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, oc)).To(Succeed())

			cond1 := apimeta.FindStatusCondition(oc.Status.Conditions, "Ready")
			Expect(cond1).NotTo(BeNil())
			Expect(cond1.Status).To(Equal(metav1.ConditionFalse))
			Expect(oc.Status.Phase).To(Or(Equal("Reconciling"), Equal("Error"), Equal("Ready")))
			Expect(oc.Status.ObservedGeneration).To(Equal(oc.Generation))

			Eventually(func(g Gomega) {
				_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
				g.Expect(err).NotTo(HaveOccurred())

				curr := &observabilityv1alpha1.ObservabilityConfig{}
				g.Expect(k8sClient.Get(ctx, typeNamespacedName, curr)).To(Succeed())

				cond := apimeta.FindStatusCondition(curr.Status.Conditions, "Ready")
				g.Expect(cond).NotTo(BeNil())
				g.Expect(cond.Status).To(Equal(metav1.ConditionTrue))
				g.Expect(curr.Status.ObservedGeneration).To(Equal(curr.Generation))
				g.Expect(curr.Status.Phase).To(Equal("Ready"))

				// Reason は共有語彙（Reconciled だけでなく DeploymentAvailable も許容）
				g.Expect(curr.Status.Reason).To(Or(
					Equal("Reconciled"),
					Equal(conditions.ReasonDeploymentAvailable),
				))
			}, 3*time.Second, 200*time.Millisecond).Should(Succeed())
		})
	})
})
