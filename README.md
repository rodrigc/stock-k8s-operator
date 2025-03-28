# 📊 Stock K8s Operator

A Kubernetes operator that dynamically monitors stock prices and integrates financial data into your Kubernetes ecosystem.

![Kubernetes + Stocks](https://raw.githubusercontent.com/rodrigc/stock-k8s-operator/main/images/stock-operator-logo.svg)

[![Go Report Card](https://goreportcard.com/badge/github.com/rodrigc/stock-k8s-operator)](https://goreportcard.com/report/github.com/rodrigc/stock-k8s-operator)
[![License](https://img.shields.io/github/license/rodrigc/stock-k8s-operator)](https://github.com/rodrigc/stock-k8s-operator/blob/main/LICENSE)

## 🔍 Overview

The Stock K8s Operator enables Kubernetes to interact with financial markets by providing real-time stock price monitoring capabilities within your cluster. This operator extends Kubernetes with custom resources that represent stock quotes and their associated data, allowing your applications to react to market information.

## ✨ Features

- 📈 **Stock Quote Monitoring**: Query current stock prices from financial data providers
- 🔄 **Automatic Updates**: Periodically fetch the latest stock information
- 📊 **Status Integration**: Stock quote data stored directly in Kubernetes resource status
- 🔗 **API Provider Flexibility**: Compatible with various stock data APIs
- 🧩 **Kubernetes Native**: Follows standard Kubernetes operator patterns

## 🏗️ Architecture

![Architecture Diagram](https://raw.githubusercontent.com/rodrigc/stock-k8s-operator/main/images/architecture.png)

The Stock K8s Operator follows standard Kubernetes operator patterns:

1. **Custom Resource Definition**: Defines `StockQuote` resources in your cluster
2. **Controller**: Watches for `StockQuote` resources and queries external APIs for data
3. **Reconciliation Loop**: Periodically fetches updated stock information
4. **Status Updates**: Records current price and metadata in the resource status

## 🚀 Installation

### Prerequisites

- Kubernetes cluster v1.19+
- kubectl v1.19+
- Helm v3+ (optional)

### Using kubectl

```bash
# Install CRDs
kubectl apply -f https://raw.githubusercontent.com/rodrigc/stock-k8s-operator/main/config/crd/bases/batch.stock-operator.crodrigues.org_stockquotes.yaml

# Install operator
kubectl apply -f https://raw.githubusercontent.com/rodrigc/stock-k8s-operator/main/config/deploy/operator.yaml
```

### Using Helm

```bash
helm repo add stock-operator https://rodrigc.github.io/stock-k8s-operator/charts
helm install stock-operator stock-operator/stock-k8s-operator
```

## 📝 Usage

### Creating a StockQuote Resource

```yaml
apiVersion: batch.stock-operator.crodrigues.org/v1
kind: StockQuote
metadata:
  name: aapl-stock
spec:
  symbol: AAPL
```

### Checking StockQuote Status

```bash
kubectl get stockquotes
```

```
NAME        SYMBOL   PRICE    LAST UPDATED
aapl-stock  AAPL     175.23   2025-03-15T14:30:00Z
msft-stock  MSFT     418.07   2025-03-15T14:30:00Z
```

### Detailed View

```bash
kubectl describe stockquote aapl-stock
```

## ⚙️ Configuration

The operator can be configured through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `WATCH_NAMESPACE` | Namespace to watch for StockQuote resources | All namespaces |
| `API_PROVIDER` | Stock data provider to use | `alphavantage` |
| `API_KEY` | API key for the stock data provider | `""` |
| `RECONCILE_PERIOD` | How often to update stock quotes | `5m` |

## 💻 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/rodrigc/stock-k8s-operator.git
cd stock-k8s-operator

# Build
make build

# Run locally
make run
```

### Running Tests

```bash
make test
```

## 🔮 Use Cases

- **Financial Dashboards**: Display real-time market data in Kubernetes-based dashboards
- **Informational Services**: Provide stock information to applications running in your cluster
- **Market Data Collection**: Gather stock prices for analysis or reporting
- **Financial Applications**: Base financial applications on current market data

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📜 License

[MIT License](LICENSE)
