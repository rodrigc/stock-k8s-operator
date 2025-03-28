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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StockQuoteSpec defines the desired state of StockQuote.
type StockQuoteSpec struct {

	// The stock ticker to retrieve (e.g. AAPL, MSFT)
	// +kubebuilder:validation:Required
	Ticker string `json:"ticker"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1440
	// TimeInterval is the interval in minutes between price updates
	TimeInterval int32 `json:"timeInterval"`

	// SecretRef refers to the secret containing the Polygon API key
	// +kubebuilder:validation:Required
	SecretRef SecretReference `json:"secretRef"`
}

// SecretReference contains details about the Secret containing the API key
type SecretReference struct {
	// Name is the name of the secret
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace containing the secret
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Key is the key in the secret containing the API key
	// +kubebuilder:default:=api-key
	Key string `json:"key,omitempty"`
}

// StockQuoteStatus defines the observed state of StockQuote.
type StockQuoteStatus struct {
	// Price is the current stock price
	Price string `json:"price,omitempty"`

	// LastUpdated is when the price was last updated
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// NextUpdateTime is when the next price update will occur
	NextUpdateTime *metav1.Time `json:"nextUpdateTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ticker",type="string",JSONPath=".spec.ticker"
// +kubebuilder:printcolumn:name="Price",type="string",JSONPath=".status.price"
// +kubebuilder:printcolumn:name="Last Updated",type="date",JSONPath=".status.lastUpdated"

// StockQuote is the Schema for the stockquotes API.
type StockQuote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StockQuoteSpec   `json:"spec,omitempty"`
	Status StockQuoteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StockQuoteList contains a list of StockQuote.
type StockQuoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StockQuote `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StockQuote{}, &StockQuoteList{})
}
