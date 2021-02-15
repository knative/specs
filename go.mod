module knative.dev/sample-source

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.2.0
	github.com/google/go-cmp v0.5.4
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.16.0
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/client-go v0.19.7
	knative.dev/eventing v0.20.1-0.20210215082944-d8468ca728b3
	knative.dev/hack v0.0.0-20210203173706-8368e1f6eacf
	knative.dev/pkg v0.0.0-20210212203835-448ae657fb5f
)

replace github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
