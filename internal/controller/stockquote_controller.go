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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	batchv1 "github.com/rodrigc/stock-k8s-operator/api/v1"
)

// StockQuoteReconciler reconciles a StockQuote object
type StockQuoteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	// APIURL is the base URL for the Polygon.io API
	APIURL string
}

type PolygonResponse struct {
	Results []struct {
		C float64 `json:"c"` // Closing price
	} `json:"results"`
}

// +kubebuilder:rbac:groups=batch.stock-operator.crodrigues.org,resources=stockquotes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.stock-operator.crodrigues.org,resources=stockquotes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch.stock-operator.crodrigues.org,resources=stockquotes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StockQuote object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *StockQuoteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling StockQuote", "namespacedName", req.NamespacedName)

	// Fetch the StockQuote instance
	stockQuote := &batchv1.StockQuote{}
	err := r.Get(ctx, req.NamespacedName, stockQuote)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted
			log.Info("StockQuote resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		log.Error(err, "Failed to get StockQuote")
		return ctrl.Result{}, err
	}

	// Check if it's time to update
	nowTime := time.Now()
	if stockQuote.Status.NextUpdateTime != nil && nowTime.Before(stockQuote.Status.NextUpdateTime.Time) {
		// Not time to update yet
		remainingTime := stockQuote.Status.NextUpdateTime.Time.Sub(nowTime)
		log.Info("Not time to update yet", "ticker", stockQuote.Spec.Ticker, "remainingTime", remainingTime)
		return ctrl.Result{RequeueAfter: remainingTime}, nil
	}

	// Get API key from the secret
	apiKey, err := r.getAPIKeyFromSecret(ctx, stockQuote)
	if err != nil {
		log.Error(err, "Failed to get API key from secret")
		return ctrl.Result{RequeueAfter: time.Minute * 5}, err
	}

	// Time to update price
	ticker := stockQuote.Spec.Ticker

	// Fetch current price from Polygon.io
	price, err := r.fetchPrice(ctx, ticker, apiKey)
	if err != nil {
		log.Error(err, "Failed to fetch price from Polygon.io", "ticker", ticker)
		return ctrl.Result{}, err
	}

	// Format price as string with 2 decimal places
	priceStr := fmt.Sprintf("%.2f", price)

	// Calculate next update time
	interval := time.Duration(stockQuote.Spec.TimeInterval) * time.Minute
	nextUpdate := nowTime.Add(interval)
	nextUpdateMeta := metav1.NewTime(nextUpdate)

	now := metav1.Now()
	// Update status
	stockQuote.Status.Price = priceStr
	stockQuote.Status.LastUpdated = &now
	stockQuote.Status.NextUpdateTime = &nextUpdateMeta

	if err := r.Status().Update(ctx, stockQuote); err != nil {
		log.Error(err, "Failed to update StockQuote status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully updated stock price",
		"ticker", ticker,
		"price", price,
		"nextUpdate", nextUpdate)

	return ctrl.Result{RequeueAfter: interval}, nil
}

// getAPIKeyFromSecret retrieves the Polygon API key from the referenced secret
func (r *StockQuoteReconciler) getAPIKeyFromSecret(ctx context.Context, stockQuote *batchv1.StockQuote) (string, error) {
	secret := &corev1.Secret{}
	secretName := stockQuote.Spec.SecretRef.Name
	secretNamespace := stockQuote.Spec.SecretRef.Namespace
	secretKey := stockQuote.Spec.SecretRef.Key

	if secretKey == "" {
		secretKey = "api-key" // Default key if not specified
	}

	err := r.Get(ctx, types.NamespacedName{
		Namespace: secretNamespace,
		Name:      secretName,
	}, secret)

	if err != nil {
		return "", fmt.Errorf("failed to get secret %s in namespace %s: %w", secretName, secretNamespace, err)
	}

	apiKey, ok := secret.Data[secretKey]
	if !ok {
		return "", fmt.Errorf("secret %s in namespace %s does not contain key %s", secretName, secretNamespace, secretKey)
	}

	return string(apiKey), nil
}

// fetchLatestPrice fetches the latest stock price from Polygon.io
func (r *StockQuoteReconciler) fetchPrice(ctx context.Context, ticker string, apiKey string) (float64, error) {
	url := fmt.Sprintf("%s/v2/aggs/ticker/%s/prev?adjusted=true&apiKey=%s", r.APIURL, ticker, apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.FromContext(ctx).Error(err, "Failed to close response body")
		}
	}()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, fmt.Errorf("rate limit exceeded for ticker: %s", ticker)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API request failed with status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var polygonResp PolygonResponse
	if err := json.NewDecoder(resp.Body).Decode(&polygonResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(polygonResp.Results) == 0 {
		return 0, fmt.Errorf("no price data available for ticker: %s", ticker)
	}

	price := polygonResp.Results[0].C
	if price <= 0 {
		return 0, fmt.Errorf("invalid price (<= 0) received for ticker: %s", ticker)
	}

	return price, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StockQuoteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.StockQuote{}).
		Named("stockquote").
		Complete(r)
}
