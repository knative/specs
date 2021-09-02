# Knative Eventing Conformance Test Plan

This document describes a plan for testing Knative Eventing Conformance based on the specs that can be found here: https://github.com/knative/specs/blob/main/specs/eventing

The specs are splitted into Control Plane and Data Plane tests, this document follows the same approach and further divide the tests into further subsections. 

# Control Plane

https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md


## Requirements: 

If you want to test conformance (**MUST, MUST NOT, REQUIRED**) you need: 
- **Prerequisites**: 
    - Knative Eventing Installed. 
    - `kubectl` access to the cluster as defined in the spec: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#rbac-profile
    - `jq` installed
- A Kubernetes Service that can be addresable to receive and count cloudevents that arrive
- `curl` to send CloudEvents

## Test Plan for Control Plane

The following sections describe the test plans for the different behaviours described in the spec and each section describe tests, commands (manual steps) and outputs that will be used to evaluate conformance. 

- [Broker Lifecycle Conformance](broker-lifecycle-conformance.md)
- [Trigger Lifecycle Conformance TBD]()
- [Channel Lifecycle Conformance TBD]()
- [Subscription Lifecycle Conformance TBD]()


# Data Plane

https://github.com/knative/specs/blob/main/specs/eventing/data-plane.md

## Test Plan for Data Plane

- [Event Ack and Delivery Retry TBD]()












