/*
Copyright 2025.

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
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	batchv1 "github.com/rodrigc/stock-k8s-operator/api/v1"
)

var _ = Describe("StockQuote Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const secretName = "test-secret"
		const secretNamespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		stockquote := &batchv1.StockQuote{}

		BeforeEach(func() {
			By("Creating the secret containing the API key")
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: secretNamespace,
				},
				Data: map[string][]byte{
					"api-key": []byte("test-api-key"),
				},
			}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: secretNamespace,
			}, secret)
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())
			}

			By("Creating the custom resource for the Kind StockQuote")
			err = k8sClient.Get(ctx, typeNamespacedName, stockquote)
			if err != nil && errors.IsNotFound(err) {
				resource := &batchv1.StockQuote{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: batchv1.StockQuoteSpec{
						Ticker:       "AAPL",
						TimeInterval: 5,
						SecretRef: batchv1.SecretReference{
							Name:      secretName,
							Namespace: secretNamespace,
							Key:       "api-key",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("Cleaning up the StockQuote resource")
			resource := &batchv1.StockQuote{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleaning up the secret")
			secret := &corev1.Secret{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: secretNamespace,
			}, secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		})

		It("should successfully reconcile the resource and update the stock price", func() {
			By("Setting up a test server to mock Polygon.io API")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"results":[{"c":150.50}]}`))
				Expect(err).NotTo(HaveOccurred())
			}))
			defer server.Close()

			By("Reconciling the created resource")
			controllerReconciler := &StockQuoteReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				APIURL: server.URL,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Verifying the stock price was updated")
			updatedStockQuote := &batchv1.StockQuote{}
			Eventually(func() error {
				return k8sClient.Get(ctx, typeNamespacedName, updatedStockQuote)
			}).Should(Succeed())

			Expect(updatedStockQuote.Status.Price).To(Equal("150.50"))
			Expect(updatedStockQuote.Status.LastUpdated).NotTo(BeNil())
			Expect(updatedStockQuote.Status.NextUpdateTime).NotTo(BeNil())
		})

		It("should handle API errors gracefully", func() {
			By("Setting up a test server to mock Polygon.io API error")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, err := w.Write([]byte(`{"error":"Internal Server Error"}`))
				Expect(err).NotTo(HaveOccurred())
			}))
			defer server.Close()

			By("Reconciling the created resource")
			controllerReconciler := &StockQuoteReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				APIURL: server.URL,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
		})
	})
})
