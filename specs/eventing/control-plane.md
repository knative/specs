# Abstract

The Knative Eventing platform provides common primitives for routing CloudEvents
between cooperating HTTP clients. This document describes the structure,
lifecycle, and management of Knative Eventing resources in the context of the
[Kubernetes Resource
Model](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/resource-management.md). An
understanding of the Kubernetes API interface and the capabilities of
[Kubernetes Custom
Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/) is assumed.

This document does not define the [data plane event delivery
contract](./data-plane.md) (though it does describe how event delivery is
configured). This document also does not prescribe specific implementations of
supporting services such as access control, observability, or resource
management.

# Background

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be
interpreted as described in [RFC 2119](https://tools.ietf.org/html/rfc2119).

There is no formal specification of the Kubernetes API and Resource Model. This
document assumes Kubernetes 1.21 behavior; this behavior will typically be
supported by many future Kubernetes versions. Additionally, this document may
reference specific core Kubernetes resources; these references may be
illustrative (i.e. _an implementation on Kubernetes_) or descriptive (i.e. _this
Kubernetes resource MUST be exposed_). References to these core Kubernetes
resources will be annotated as either illustrative or descriptive.

This document considers two users of a given Knative Eventing environment, and is
particularly concerned with the expectations of developers (and language and
tooling developers, by extension) deploying applications to the environment.

- **Developers** configure Knative resources to implement an event-routing
  architecture.
- **Operators** (also known as **platform providers**) provision the underlying
  event routing resources and manage the software configuration of Knative
  Eventing and the underlying abstractions.

# RBAC Profile

In order to validate the controls described in [Resource
Overview](#resource-overview), the following Kubernetes RBAC profile may be
applied in a Kubernetes cluster. This Kubernetes RBAC is an illustrative example
of the minimal profile rather than a requirement. This Role should be sufficient
to develop, deploy, and manage event routing for an application within a single
namespace. Knative Conformance tests against "MUST", "MUST NOT", and "REQUIRED"
conditions are expected to pass when using this profile:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: knative-developer
rules:
  - apiGroups: ["eventing.knative.dev"]
    resources: ["broker", "trigger"]
    verbs: ["get", "list", "create", "update", "delete"]
  - apiGroups: ["messaging.knative.dev"]
    resources: ["channel", "subscription"]
    verbs: ["get", "list", "create", "update", "delete"]
```

<!-- TODO: define aggregated API roles

Ref:
- https://github.com/knative/specs/blob/main/specs/eventing/channel.md#aggregated-channelable-manipulator-clusterrole
- https://github.com/knative/specs/blob/main/specs/eventing/channel.md#aggregated-addressable-resolver-clusterrole
- 
-->

# Resource Overview

The Knative Eventing API provides a set of primitives to support both
point-to-point communication channels (`messaging.knative.dev`) and
content-based event routing (`eventing.knative.dev`). This specification
describes API interfaces of Knative Eventing resources as well as the supported
[event routing](./data-plane.md) logic and configuration settings.

At the moment, the Knative Eventing specification does not contemplate any
non-Kubernetes-backed implementations, and therefore does not specifically
define the mapping of kubernetes verbs (read, watch, patch, etc) to developer
roles. See the [Overview documentation](./overview.md) for general definitions
of the different API objects.

# Error Signalling
<!-- copied from ../serving/knative-api-specification-1.0.md#error-signalling -->

The Knative API uses the
[Kubernetes Conditions convention](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)
to communicate errors and problems to the user. Each user-visible resource
described in Resource Overview MUST have a `conditions` field in `status`, which
must be a list of `Condition` objects of the following form (note that the
actual API object types may be named `FooCondition` to allow better code
generation and disambiguation between similar fields in the same `apiGroup`):

<table>
  <tr>
   <td><strong>Field</strong>
   </td>
   <td><strong>Type</strong>
   </td>
   <td><strong>Description</strong>
   </td>
   <td><strong>Default Value</strong>
   </td>
  </tr>
  <tr>
   <td><code>type</code>
   </td>
   <td><code>string</code>
   </td>
   <td>The category of the condition, as a short, CamelCase word or phrase.
<p>
This is the primary key of the Conditions list when viewed as a map.
   </td>
   <td>REQUIRED – No default
   </td>
  </tr>
  <tr>
   <td><code>status</code>
   </td>
   <td>Enum:<ul>

<li>"True"
<li>"False"
<li>"Unknown"</li></ul>

   </td>
   <td>The last measured status of this condition.
   </td>
   <td>"Unknown"
   </td>
  </tr>
  <tr>
   <td><code>reason</code>
   </td>
   <td>string
   </td>
   <td>One-word CamelCase reason for the condition's last transition.
   </td>
   <td>""
   </td>
  </tr>
  <tr>
   <td><code>message</code>
   </td>
   <td>string
   </td>
   <td>Human-readable sentence describing the last transition.
   </td>
   <td>""
   </td>
  </tr>
  <tr>
   <td><code>severity</code>
   </td>
   <td>Enum:<ul>

<li>""
<li>"Warning"
<li>"Info"</li></ul>

   </td>
   <td>If present, represents the severity of the condition. An empty severity represents a severity level of "Error". 
   </td>
   <td>""
   </td>
  </tr>
  <tr>
   <td><code>lastTransitionTime</code>
   </td>
   <td>Timestamp
   </td>
   <td>Last update time for this condition.
   </td>
   <td>"" – may be unset
   </td>
  </tr>
</table>

Additionally, the resource's `status.conditions` field MUST be managed as
follows to enable clients (particularly user interfaces) to present useful
diagnostic and error message to the user. In the following section, conditions
are referred to by their `type` (aka the string value of the `type` field on the
Condition).

1.  Each resource MUST have either a `Ready` condition (for ongoing systems) or
    `Succeeded` condition (for resources that run to completion) with
    `severity=""`, which MUST use the `True`, `False`, and `Unknown` status
    values as follows:

    1.  `False` MUST indicate a failure condition.
    1.  `Unknown` SHOULD indicate that reconciliation is not yet complete and
        success or failure is not yet determined.
    1.  `True` SHOULD indicate that the application is fully reconciled and
        operating correctly.

    `Unknown` and `True` are specified as SHOULD rather than MUST requirements
    because there may be errors which prevent serving which cannot be determined
    by the API stack (e.g. DNS record configuration in certain environments).
    Implementations are expected to treat these as "MUST" for factors within the
    control of the implementation.

1.  For non-`Ready` conditions, any conditions with `severity=""` (aka "Error
    conditions") must be aggregated into the "Ready" condition as follows:

    1.  If the condition is `False`, `Ready` MUST be `False`.
    1.  If the condition is `Unknown`, `Ready` MUST be `False` or `Unknown`.
    1.  If the condition is `True`, `Ready` may be any of `True`, `False`, or
        `Unknown`.

    Implementations MAY choose to report that `Ready` is `False` or `Unknown`
    even if all Error conditions report a status of `True` (i.e. there may be
    additional hidden implementation conditions which feed into the `Ready`
    condition which are not reported.)

1.  Non-`Ready` conditions with non-error severity MAY be surfaced by the
    implementation. Examples of `Warning` or `Info` conditions could include:
    missing health check definitions, scale-to-zero status, or non-fatal
    capacity limits.

Conditions type names should be chosen to describe positive conditions where
`True` means that the condition has been satisfied. Some conditions may be
transient (for example, `ResourcesAllocated` might change between `True` and
`False` as an application scales to and from zero). It is RECOMMENDED that
transient conditions be indicated with a `severity="Info"`.

# Resource Lifecycle

## Broker

TODO: lifecycle; are triggers deleted when the broker is deleted? Are trigger conditions reflected in the broker? What are the required conditions (Ready, ???)? `spec.class` SHOULD be immutable.

## Trigger

TODO: lifecycle; what happens if a trigger is created before a broker? What about events received by a broker before the trigger is created? Are broker conditions reflected in the trigger? What are the required conditions (Ready, ???)? `spec.broker` SHOULD be immutable.

## Channel

TODO: lifecycle; are subscriptions deleted when the channel is deleted? Are subscription conditions reflected in the channel? What are the required conditions (Ready, ???)?  <!-- should `spec.channelTemplate` be immutable? -->

TODO: channel-compatible CRDs (Channelable)

## Subscription

TODO: lifecycle; what happens if a subscription is created before a channel?  What about events received by a channel before the subscription is created? Are channel conditions reflected in the subscription? What are the required conditions (Ready, ???)? `spec.channel` SHOULD be immutable.

## Addressable resolution

TODO: What is it? How does it apply to Trigger / Subscription? Indicate that Broker & Channel MUST implement Addressable.

# Event Routing

## Content Based Routing

TODO: How do Broker & Trigger handle routing.
- When must events arriving at a Broker be routed by a Trigger (when the Trigger is Ready?)
  - How do retries and replies interact with configuration changes in the Trigger / Broker?
- Retry parameters
- Reply routing
- Dead-letter routing

## Topology Based Routing

TODO: How do Channel & Subscription handle routing
- When must events arriving at a Channel be routed to a Subscription (when the Subscription is Ready?)
  - How do retries and replies interact with configuration changes in the Subscription / Channel?
- Retry parameters
- Reply routing
- Dead-letter routing

## Event Sources

TODO: What's required of an event source with respect to routing? Retries? Dead-letter? Status?

# Detailed Resources

## Broker v1

## Trigger v1

## Channel v1

## Subscription v1

## Addressable v1

================================================================
    CUT HERE
================================================================

## Broker

The Knative Broker represents a single instance of an event router which accepts
events from one or more sources and routes them to selected destinations based
on rules matching the attributes of the received event. In order to do this, the Broker defines 
