# Knative Eventing API spec

This directory contains the specification of the Knative Eventing API, which is
implemented in [`eventing.knative.dev`](https://github.com/knative/eventing/blob/main/pkg/apis/eventing/v1) and verified
via [the e2e test](https://github.com/knative/eventing/blob/main/test/e2e).

**Updates to this spec should include a corresponding change to the API
implementation for [eventing](https://github.com/knative/eventing/blob/main/pkg/apis/eventing/v1beta1) and
[the e2e test](https://github.com/knative/eventing/blob/main/test/e2e).**

Docs in this directory:

- [Motivation and goals](motivation.md)
- [Resource type overview](overview.md)
- [Control Plane specification](control-plane.md)
- [Data Plane specification](data-plane.md)
