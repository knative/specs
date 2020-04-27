module knative.dev/sample-source

go 1.13

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.1 // indirect
	contrib.go.opencensus.io/exporter/zipkin v0.1.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/cloudevents/sdk-go/v2 v2.0.0-RC1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	go.uber.org/atomic v1.5.1 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/zap v1.13.0
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.0.0-00010101000000-000000000000 // indirect
	k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.0 // indirect
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	knative.dev/eventing v0.14.1-0.20200427112650-40f0a540923e
	knative.dev/pkg v0.0.0-20200427190051-6b9ee63b4aad
	knative.dev/test-infra v0.0.0-20200424202250-e6e89d29e93a // indirect
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/code-generator => k8s.io/code-generator v0.16.4
)
