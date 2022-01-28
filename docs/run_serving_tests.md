# Running Knative Serving Conformance Tests

## Tests

The standard set of conformance tests is currently those under the [Knative
serving conformance directory](https://github.com/knative/serving/tree/main/test/conformance).

## Instructions

1. Check out your Knative Serving fork

    Create your own fork of the Knative Serving repository. Clone it to your machine:

    ```sh
    git clone git@github.com:<YOUR_GITHUB_USERNAME>/serving.git
    cd serving
    ```

1. Build and upload the test images

    Set the environment variable for the docker repository to which test images should be pushed (e.g. gcr.io/[gcloud-project]).
    Then build and upload the test images

    ```sh
    export KO_DOCKER_REPO=<YOUR_DOCKER_REPOSITORY>
    ./test/upload-test-images.sh
    ```

1. Configure client objects

    By default the tests will use the
    [kubeconfig file](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
    at `~/.kube/config` to create [client objects](https://github.com/knative/serving/blob/ff30afc5fa738236181f50bc2e480061ad5a097d/test/clients.go#L45)
    connecting to your Knative platform. If `~/.kube/config` is not the correct
    config file for you to use, create your own.

    You might need take other actions to grant authorization for the client objects to access
    your Knative platform.

    Please see [examples](#examples) section for details.

1. Run the tests

    Run conformance tests for API v1:

    ```sh
    go test -v -tags=e2e -count=1 ./test/conformance/api/v1 \
    -kubeconfig <YOUR_CONFIG_FILE> -disable-logstream=true -disable-optional-api=true
    ```

    Run conformance tests for runtime:

    ```sh
    go test -v -tags=e2e -count=1 ./test/conformance/runtime \
    -kubeconfig <YOUR_CONFIG_FILE> -disable-logstream=true -disable-optional-api=true
    ```

    You might need to specify extra [test flags](https://github.com/knative/serving/blob/main/test/e2e_flags.go).

## Examples

Examples are under this [directory](./examples/serving/).
