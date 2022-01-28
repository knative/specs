# Running Knative Serving Conformance Tests against Google Cloud Run

1. Check out your Knative Serving fork

Create your own fork of the Knative Serving repository. Clone it to your machine:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/serving.git
cd serving
```

1. Build and upload the test images

```sh
export KO_DOCKER_REPO='gcr.io/my-gcp-project'
./test/upload-test-images.sh
```

1. Configure client objects

Save the following content to a file, for example `~/kubeconfig_for_cloud_run`.

```yaml
apiVersion: v1
clusters:
- cluster:
    server: https://us-central1-run.googleapis.com
  name: hosted
contexts:
- context:
    cluster: hosted
    user: hosteduser
  name: hostedcontext
current-context: hostedcontext
kind: Config
preferences: {}
users:
- name: hosteduser
  user:
    auth-provider:
      config: null
      name: gcp
```

This file configures the client objects to connect to Cloud Run server in region `us-central1`.
Change this value if you’d like to use a different [region](https://cloud.google.com/run/docs/locations).

Make sure your [gcloud](https://cloud.google.com/sdk/docs/install) account has the Cloud Run
Admin role ([roles/run.admin](https://cloud.google.com/run/docs/reference/iam/roles)).
Authorize your gcloud account credentials:

```sh
gcloud auth login
gcloud auth application-default login
```

1. Run the tests

Run conformance tests for API v1:

```sh
go test -v -tags=e2e -count=1 ./test/conformance/api/v1 \
 -kubeconfig ~/kubeconfig_for_cloud_run -disable-logstream=true -disable-optional-api=true \
 -resolvabledomain=true -test-namespace=my-gcp-project \
 -request-headers="Authorization,Bearer $(gcloud auth print-identity-token)" \
 -exceeding-memory-limit-size=10000 
```

Run conformance tests for runtime:

```sh
go test -v -tags=e2e -count=1 ./test/conformance/runtime \
  -kubeconfig ~/kubeconfig_for_cloud_run -disable-logstream=true -disable-optional-api=true \
  -resolvabledomain=true -test-namespace=my-gcp-project \
  -request-headers="Authorization,Bearer $(gcloud auth print-identity-token)"
```

Explanations for some flags:

- `kubeconfig`: The kubeconfig file you created in the previouse step.

- `disable-optional-api`: Skip the tests against optional APIs. The optional APIs are not required
  by Knative Specification. Cloud Run does not support some of them.

- `resolvabledomain`: The URL for a Cloud Run service route is resolvable.

- `test-namespace`: Cloud Run uses Google Cloud Platform project ID as the namespace.

- `request-headers`: Identity token from gcloud is used as authorization header for testing requests to get access to Cloud Run services.

- `exceeding-memory-limit-size`: [Knative runtime contract](https://github.com/knative/specs/blob/main/specs/serving/runtime-contract.md#memory-and-cpu-limits)
  allows the serverless platform to automatically adjust the resource limits (e.g. memory) based on
  observed resource usage. Cloud Run does so. When a memory exceeding usage happens, Cloud Run
  
  - MAY adjust the limit to a much higher value temporarily to handle a memory
    usage peak for user applications.

  - sends a warning log to ask the user to increase the memory limit.

  10GB is used here instead of the default 500MB in order to get a non-200 response due to memory usage exceeding the 300MB limit.