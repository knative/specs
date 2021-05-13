# Abstract

The Knative Eventing platform provides common primitives for routing CloudEvents
between cooperating HTTP clients. This document describes the structure,
lifecycle, and management of Knative Eventing resources in the context of the
[Kubernetes Resource Model](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture/resource-management.md).
An understanding of the Kubernetes API interface and the capabilities of
[Kubernetes Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
is assumed.

This document does not define the
[data plane event delivery contract](./data-plane.md) (though it does describe
how event delivery is configured). This document also does not prescribe
specific implementations of supporting services such as access control,
observability, or resource management.

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

This document considers two users of a given Knative Eventing environment, and
is particularly concerned with the expectations of developers (and language and
tooling developers, by extension) deploying applications to the environment.

- **Developers** configure Knative resources to implement an event-routing
  architecture.
- **Operators** (also known as **platform providers**) provision the underlying
  event routing resources and manage the software configuration of Knative
  Eventing and the underlying abstractions.

# RBAC Profile

In order to validate the controls described in
[Resource Overview](#resource-overview), the following Kubernetes RBAC profile
may be applied in a Kubernetes cluster. This Kubernetes RBAC is an illustrative
example of the minimal profile rather than a requirement. This Role should be
sufficient to develop, deploy, and manage event routing for an application
within a single namespace. Knative Conformance tests against "MUST", "MUST NOT",
and "REQUIRED" conditions are expected to pass when using this profile:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: knative-developer
rules:
  - apiGroups: ["eventing.knative.dev"]
    resources: ["broker", "trigger"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: ["messaging.knative.dev"]
    resources: ["channel", "subscription"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
```

In order to support resolving resources which meet the
[Addressable](./overview.md#addressable) contract, the system controlling the
[Trigger](#trigger-lifecycle) or [Subscription](#subscription-lifecycle) will
need _read_ access to these resources. On Kubernetes, this is most easily
achieved using role aggregation; on systems using Kubernetes RBAC, resources
which wish to participate in Addressable resolution should provide the following
`ClusterRole`:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: <component>-addressable-resolver
  labels:
    duck.knative.dev/addressable: "true"
rules:
  - apiGroups: [<resource apiGroup>]
    resources: [<resource kind >, <resource kind>/status]
    verbs: ["get", "list", "watch"]
```

This configuration advice SHALL NOT indicate a requirement for Kubernetes RBAC
support.

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
    because there may be errors which prevent functioning which cannot be
    determined by the API stack (e.g. DNS record configuration in certain
    environments). Implementations are expected to treat these as "MUST" for
    factors within the control of the implementation.

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

## Broker Lifecycle

A Broker represents an Addressable endpoint (i.e. it has a `status.address.url`
field) which can receive, store, and forward events to multiple recipients based
on a set of attribute filters (Triggers). Triggers are associated with a Broker
based on the `spec.broker` field on the Trigger; it is expected that the
controller for a Broker will also control the associated Triggers. When the
Broker's `Ready` condition is `true`, the Broker MUST provide a
`status.address.url` which accepts all valid CloudEvents and MUST forward the
received events for filtering to each associated Trigger whose `Ready` condition
is `true`. As described in the [Trigger Lifecycle](#trigger-lifecycle) section,
a Broker MAY forward events to an associated Trigger which which does not
currently have a `true` `Ready` condition, including events received by the
Broker before the Trigger was created.

The annotation `eventing.knative.dev/broker.class` may be used to select a
particular implementation of a Broker. When a Broker is created, its
implementation and the `spec.config` field MUST be populated (`spec.config` MAY
be an empty broker) to indicate which of several possible Broker implementations
to use. It is RECOMMENDED to default the `eventing.knative.dev/broker.class`
field on creation if it is unpopulated. Once created, both fields MUST be
immutable; the Broker must be deleted and re-created to change the
implementation class or `spec.config`. This pattern is chosen to make it clear
that changing the implementation class or `spec.config` is not an atomic
operation and that any implementation would be likely to result in event loss
during the transition.

## Trigger Lifecycle

The lifecycle of a Trigger is independent of the Broker it refers to in its
`spec.broker` field; if the Broker does not currently exist or the Broker's
`Ready` condition is not `true`, then the Trigger's `Ready` condition MUST be
`false`, and the reason SHOULD indicate that the corresponding Broker is missing
or not ready.

The Trigger MUST also set the `status.subscriberUri` field based on resolving
the `spec.subscriber` field before setting the `Ready` condition to `true`. If
the `spec.subscriber.ref` field points to a resource which does not exist or
cannot be resolved via [Addressable resolution](#addressable-resolution), the
Trigger MUST set the `Ready` condition to `false`, and at least one condition
should indicate the reason for the error.

If the Trigger's `spec.delivery.deadLetterSink` field it set, it MUST be
resolved to a URI and reported in `status.deadLetterSinkUri` in the same manner
as the `spec.subscriber` field before setting the `Ready` condition to `true`.

Once created, the Trigger's `spec.broker` SHOULD NOT permit updates; to change
the `spec.broker`, the Trigger can be deleted and re-created. This pattern is
chosen to make it clear that changing the `spec.broker` is not an atomic
operation, as it may span multiple storage systems. Changes to
`spec.subscriber`, `spec.filter` and other fields SHOULD be permitted, as these
could occur within a single storage system.

When a Trigger becomes associated with a Broker (either due to creating the
Trigger or the Broker), the Trigger MUST only set the `Ready` condition to
`true` after the Broker has been configured to send all future events matching
the `spec.filter` to the Trigger's `spec.subscriber`. The Broker MAY send some
events to the Trigger's `spec.subscriber` prior to the Trigger's
`Ready`condition being set to `true`. When a Trigger is deleted, the Broker MAY
send some additional events to the Trigger's `spec.subscriber` after the
deletion.

## Channel Lifecycle

A Channel represents an Addressable endpoint (i.e. it has as
`status.address.url` field) which can receive, store, and forward events to
multiple recipients (Subscriptions). Subscriptions are associated with a Channel
based on the `spec.channel` field on the Subscription; it is expected that the
controller for a Channel will also control the associated Subscriptions. When
the Channel's `Ready` condition is `true`, the Channel MUST provide a
`status.address.url` which accepts all valid CloudEvents and MUST forward the
received events to each associated Subscription whose `Ready` condition is
`true`. As described in the [Subscription Lifecycle](#subscription-lifecycle)
section, a Channel MAY forward events to an associated Subscription which does
not currently have a `true` `Ready` condition, including events received by the
Channel before the `Subscription` was created.

When a Channel is created, its `spec.channelTemplate` field MUST be populated to
indicate which of several possible Channel implementations to use. It is
RECOMMENDED to default the `spec.channelTemplate` field on creation if it is
unpopulated. Once created, the `spec.channelTemplate` field MUST be immutable;
the Channel MUST be deleted and re-created to change the `spec.channelTemplate`.
This pattern is chosen to make it clear that changing `spec.channelTemplate` is
not an atomic operation and that any implementation would be likely to result in
event loss during the transition.

## Subscription Lifecycle

The lifecycle of a Subscription is independent of that of the channel it refers
to in its `spec.channel` field. The `spec.channel` object reference may refer to
either an `messaging.knative.dev/v1` Channel resource, or another resource which
meets the `spec.subscribers` and `spec.delivery` required elements in the
Channelable duck type. If the referenced `spec.channel` does not currently exist
or its `Ready` condition is not `true`, then the Subscription's `Ready`
condition MUST NOT be `true`, and the reason SHOULD indicate that the
corresponding channel is missing or not ready.

The Subscription MUST also set the `status.physicalSubscription` URIs by
resolving the `spec.subscriber`, `spec.reply`, and
`spec.delivery.deadLetterSink` as described in
[Addressable resolution](#addressable-resolution) before setting the `Ready`
condition to `true`. If any of the addressable fields fails resolution, the
Subscription MUST set the `Ready` condition to `false`, and at least one
condition should indicate the reason for the error. (It is acceptable for none
of the `spec.subscriber`, `spec.reply`, and `spec.delivery.deadLetterSink`
fields to contain a `ref` field.)

At least one of `spec.subscriber` and `spec.reply` MUST be set; if only
`spec.reply` is set, the behavior is equivalent to setting `spec.subscriber`
except that the Channel SHOULD NOT advertise the ability to process replies
during the delivery.

Once created, the Subscription's `spec.channel` SHOULD NOT permit updates; to
change the `spec.channel`, the Subscription can be deleted and re-created. This
pattern is chosen to make it clear that changing the `spec.channel` is not an
atomic operation, as it may span multiple storage systems. Changes to
`spec.subscriber`, `spec.reply`, `spec.delivery` and other fields SHOULD be
permitted, as these could occur within a single storage system.

When a Subscription becomes associated with a channel (either due to creating
the Subscription or the channel), the Subscription MUST only set the `Ready`
condition to `true` after the channel has been configured to send all future
events to the Subscriptions `spec.subscriber`. The Channel MAY send some events
to the Subscription before prior to the Subscription's `Ready` condition being
set to `true`. When a Subscription is deleted, the Channel MAY send some
additional events to the Subscription's `spec.subscriber` after the deletion.

<!--
TODO: channel-compatible CRDs (Channelable)
-->

## Addressable Resolution

Both Trigger and Subscription have optional object references (`ref` in
`spec.subscriber`, `spec.delivery.deadLetterSink`, and `spec.reply` for
Subscription) which are expected to conform to the Addressable partial schema
("duck type"). An object conforms with the Addressable partial schema if it
contains a `status.address.url` field containing a URL which may be used to
deliver CloudEvents over HTTP. As a special case, Kubernetes `v1` `Service`
objects are considered to have a `status.address.url` of
`http://<service-dns-name>/`.

If both the `ref` field and the `uri` fields are set on a Destination, then the
Destination's address is the `uri` interpreted relative to the resolved URI of
the `ref` field. This may be used, for example, to refer to a specific URL off a
referenced domain name, like so:

```yaml
subscriber:
  ref:
    apiVersion: v1
    kind: Service
    name: test
  uri: "/update
```

This Destination would resolve to
`http://test.<current-namespace-dns-prefix>/update`, as `/update` would be
interpreted relative to the `test` Service's DNS name.

If one of these object references points to an object which does not currently
satisfy this partial schema (either because the `status.address.url` field is
empty, or because the object does not have that field), then the Trigger or
Subscription MUST indicate an error by setting the `Ready` condition to `false`,
and SHOULD include an indication of the error in a condition reason or type.

Both Broker and Channel MUST conform to the Addressable partial schema.

# Event Routing

Note that the event routing description below does not cover the actual
mechanics of sending an event from one component to another; see
[the data plane](./data-plane.md) contracts for details of the event transfer
mechanisms.

## Content Based Routing

A Broker MUST publish a URL at `status.address.uri` when it is able to receive
events. This URL MUST accept CloudEvents in both the
[Binary Content Mode](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#31-binary-content-mode)
and
[Structured Content Mode](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#32-structured-content-mode)
HTTP formats. Before sending an HTTP response, the Broker MUST durably enqueue
the event (be able to deliver with retry without receiving the event again).

For each event received by the Broker, the Broker MUST evaluate each associated
Trigger **once** (where "associated" means a Trigger with a `spec.broker` which
references the Broker). If the Trigger has a `Ready` condition of `true` when
the event is evaluated, the the Broker MUST evaluate the Trigger's `spec.filter`
and, if matched, proceed with
[event delivery as described below](#event-delivery). The Broker MAY also
evaluate and forward events to associated Triggers for which the `Ready`
condition is not currently `true`. (One example: a Trigger which is in the
process of being programmed in the Broker data plane might receive _some_ events
before the data plane programming was complete and the Trigger was updated to
set the `Ready` condition to `true`.)

If multiple Triggers match an event, one event delivery MUST be generated for
each match; duplicate matches with the same destination MUST each generate a
separate event delivery attempts, one per Trigger match. The implementation MAY
attach additional event attributes or other metadata distinguishing between these
deliveries. The implementation MUST NOT modify the event payload in this
process.

Reply events generated during event delivery MUST be re-enqueued by the Broker
in the same way as events delivered to the Broker's Addressable URL. If the
storage of the reply event fails, the entire event delivery MUST be failed and
the delivery to the Trigger's subscriber MUST be retried. Reply events re-enqueued
in this manner MUST be evaluated against all triggers associated with the
Broker, including the Trigger that generated the reply. Implementations MAY
implement event-loop detection; it is RECOMMENDED that any such controls be
documented to end-users. Implementations MAY avoid using HTTP to deliver event
replies to the Broker's event-delivery input and instead use an internal
queueing mechanism.

## Topology Based Routing

A Channel MUST publish a URL at `status.address.url` when it is able to receive
events. This URL MUST accept CloudEvents in both the
[Binary Content Mode](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#31-binary-content-mode)
and
[Structured Content Mode](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#32-structured-content-mode)
HTTP formats. Before sending an HTTP response, the Channel MUST durably enqueue
the event (be able to deliver with retry without receiving the event again).

For each event received by the Channel, the Channel MUST deliver the event to
each associated Subscription **at least once** (where "associated" means a
Subscription with a `spec.channel` which references the Channel). If the
Subscription has a `Ready` condition of `true` when the event is evaluated, the
Channel MUST forward the event as described in
[event delivery as described below](#event-delivery). The Channel MAY also
forward events to associated Subscriptions for with the `Ready` condition is not
currently `true`. (One example: a Subscription which is in the process of being
programmed in the Channel data plane might receive _some_ events before the data
plane programming was complete and the Subscription was updated to set the
`Ready` condition to `true`.)

If multiple Subscriptions with the same destination are associated with the same
Channel, each Subscription MUST generate one delivery attempt per Subscription.
The implementation MAY attach additional event attributes or other metadata
distinguishing between these deliveries. The implementation MUST NOT modify the
event payload in this process.

## Event Delivery

Once a Trigger or Subscription has decided to deliver an event, it MUST do the
following:

1. Resolve all URLs and delivery options, using the values in `status` for URL
   resolution.

1. Attempt delivery to the `status.subscriberUri` URL following the
   [data plane contract](./data-plane.md).

   1. If the event delivery fails with a retryable error, it SHOULD be retried
      up to `retry` times, following the `backoffPolicy` and `backoffDelay`
      parameters if specified.

1. If the delivery attempt is successful (either the original request or a
   retry) and no event is returned, the event delivery is complete.

1. If the delivery attempt is successful (either the original request or a
   retry) and an event is returned in the reply, the reply event will be
   delivered to the `status.replyUri` destination (for Subscriptions) or added
   to the Broker for processing (for Triggers). If `status.replyUri` is not
   present in the Subscription, the reply event is be dropped.

   1. If delivery of the reply event fails with a retryable error, the delivery
      of the reply event to the reply destination SHOULD be retried up to
      `retry` times, following the `backoffPolicy` and `backoffDelay` parameters
      if specified.

1. If all retries are exhausted for either the original delivery or the retry,
   or if a non-retryable error is received, the event should be delivered to the
   `deadLetterSink` in the delivery options. If no `deadLetterSink` is
   specified, the event is dropped.

   The implementation MAY set additional attributes on the event or wrap the
   failed event in a "failed delivery" event; this behavior is not (currently)
   standardized.

   If delivery of the reply event fails with a retryable error, the delivery to
   the `deadLetterSink` SHOULD be retried up to `retry` times, following the
   `backoffPolicy` and `backoffDelay` parameters if specified. Alternatively,
   implementations MAY use an equivalent internal mechanism for delivery (for
   example, if the `ref` form of `deadLetterSink` points to a compatible
   implementation).

# Detailed Resources

## Broker v1

### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource.

### Spec:

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>config</code></td>
    <td>KReference<br/>(Optional)</td>
    <td>A reference to an object which describes the configuration options for the Broker (for example, a ConfigMap).</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td>DeliverySpec<br/>(Optional)</td>
    <td>A default delivery options for Triggers which do not specify more-specific options.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link. Conditions of types Ready MUST be present.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest metadata.generation that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td>duckv1.Addressable</td>
    <td>Address used to deliver events to the Broker.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</td>
    <td>URL (string)</td>
    <td>If <code>spec.delivery.deadLetterSink</code> is specified, the resolved URL of the dead letter address.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## Trigger v1

### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource.

### Spec:

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>broker</code></td>
    <td>string<br/>(Required, Immutable)</td>
    <td>The Broker which this Trigger should filter events from.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>filter</code></td>
    <td>TriggerFilter<br/>(Optional)</td>
    <td>Event filters which are used to select events to be delivered to the Trigger's destination.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>subscriber</code></td>
    <td>duckv1.Destination<br/>(Required)</td>
    <td>The destination that filtered events should be sent to.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td>DeliverySpec<br/>(Optional)</td>
    <td>Delivery options for this Trigger.</td>
    <td>RECOMMENDED</td>
  </tr>
</table>

### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link. Conditions of types Ready MUST be present.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest metadata.generation that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>subscriberUri</code></td>
    <td>URL (string)</td>
    <td>The resolved address of the <code>spec.subscriber</code>.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</td>
    <td>URL (string)</td>
    <td>If <code>spec.delivery.deadLetterSink</code> is specified, the resolved URL of the dead letter address.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## Channel v1

### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource.

### Spec:

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>channelTemplate</code></td>
    <td>object<br/>(Optional)</td>
    <td>Implementation-specific parameters to configure the channel.</td>
    <td>OPTIONAL</td>
  </tr>
  <tr>
    <td><code>subscribers</code></td>
    <td>[]duckv1.SubscriberSpec</td>
    <td>Aggregated subscription information; this array should be managed automatically by the controller.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td>DeliverySpec<br/>(Optional)</td>
    <td>A default delivery options for Subscriptions which do not specify more-specific options.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link. Conditions of types Ready MUST be present.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest metadata.generation that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td>duckv1.Addressable</td>
    <td>Address used to deliver events to the Broker.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>subscribers</code></td>
    <td>[]duckv1.SubscriberStatus</td>
    <td>Resolved addresses for the <code>spec.subscribers</code> (subscriptions to this Channel).</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</td>
    <td>URL (string)</td>
    <td>If <code>spec.delivery.deadLetterSink</code> is specified, the resolved URL of the dead letter address.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## Subscription v1

### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource.

### Spec:

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>channel</code></td>
    <td>Kubernetes v1/ObjectReference<br/>(Required, Immutable)</td>
    <td>The channel this subscription receives events from. Immutable.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>subscriber</code></td>
    <td>duckv1.Destination<br/>(Required)</td>
    <td>The destination that events should be sent to.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>reply</code></td>
    <td>duckv1.Destination<br/>(Required)</td>
    <td>The destination that reply events from <code>spec.subscriber</code> should be sent to.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td>DeliverySpec<br/>(Optional)</td>
    <td>Delivery options for this Subscription.</td>
    <td>RECOMMENDED</td>
  </tr>
</table>

### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link. Conditions of types Ready MUST be present.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest metadata.generation that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>physicalSubscription</code></td>
    <td>PhysicalSubscriptionStatus</td>
    <td>The fully resolved values for <code>spec</code> endpoint references.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## Addressable v1

Note that the Addressable interface is a partial schema -- any resource which
includes these fields may be referenced using a `duckv1.Destination`.

### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource.

### Spec:

There are no `spec` requirements for Addressable.

### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td>duckv1.Addressable</td>
    <td>Address used to deliver events to the Broker.</td>
    <td>REQUIRED</td>
  </tr>
</table>

# Detailed SubResource Objects

## duckv1.Addressable

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>url</code></td>
    <td>URL</td>
    <td>Address used to deliver events to the Addressable.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## duckv1.Destination

Destination is used to indicate the destination for event delivery. A
Destination eventually resolves the supplied information to a URL by resolving
`uri` relative to the address of `ref` (if provided). `ref` may be an
[Addressable](#duckv1-addressable) object or a `v1/Service` Kubernetes service.

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>ref</code></td>
    <td>duckv1.KReference<br/>(Optional)</td>
    <td>An ObjectReference to an Addressable reference to deliver events to.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>uri</code></td>
    <td>URL (string)<br/>(Optional)</td>
    <td>A resolved URL to deliver events to.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## duckv1.SubscriberSpec

SubscriberSpec represents an automatically-populated extraction of information
from a [Subscription](#subscription).

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>uid</code></td>
    <td>UID (string)</td>
    <td>UID is used to disambiguate Subscriptions which might be recreated.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>generation</code></td>
    <td>int64</td>
    <td>Generation of the copied Subscription.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>subscriberUri</code></td>
    <td>URL (string)</td>
    <td>The resolved address of the Subscription's <code>spec.subscriber</code>.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>replyUri</code></td>
    <td>URL (string)</td>
    <td>The resolved address of the Subscription's <code>spec.reply</code>.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td>DeliverySpec</td>
    <td>The resolved Subscription delivery options. The <code>deadLetterSink</code> should use the <code>uri</code> form.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## duckv1.SubscriberStatus

SubscriberStatus indicates the status of programming a Subscription by a
Channel.

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>uid</code></td>
    <td>UID (string)</td>
    <td>UID is used to disambiguate Subscriptions which might be recreated.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>generation</code></td>
    <td>int64</td>
    <td>Generation of the copied Subscription.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>ready</code></td>
    <td>kubernetes v1/ConditionStatus</td>
    <td>Ready status of the Subscription's programming into the Channel data plane.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>message</code></td>
    <td>duckv1.Addressable</td>
    <td>A human readable message indicating details of <code>ready</code> status.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## DeliverySpec

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>deadLetterSink</code></td>
    <td>duckv1.Destination</td>
    <td>Fallback address used to deliver events which cannot be delivered during the flow. An implementation MAY place limits on the allowed destinations for the <code>deadLetterSink</code>.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>retry</code></td>
    <td>int</td>
    <td>Retry is the minimum number of retries the sender should attempt when sending an event before moving it to the dead letter sink.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>backoffDelay</code></td>
    <td>string</td>
    <td>The initial delay when retrying delivery.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>backoffPolicy</code></td>
    <td>enum<br/>("linear", "exponential")</td>
    <td>Retry timing scaling policy. Linear policy uses the same <code>backoffDelay</code> for each attempt; Exponential policy uses 2^N multiples of <code>backoffDelay</code></td>
    <td>RECOMMENDED</td>
  </tr>
</table>

## KReference

KReference is a lightweight version of kubernetes
[`v1/ObjectReference`](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectreference-v1-core)

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>apiVersion</code></td>
    <td>string</td>
    <td>ApiVersion of the target reference.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>kind</code></td>
    <td>string</td>
    <td>Kind of the target reference.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>namespace</code></td>
    <td>string</td>
    <td>Namespace of the target resource.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td>string</td>
    <td>Address used to deliver events to the Broker.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## PhysicalSubscriptionStatus

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>subscriberUri</code></td>
    <td>URL (string)</td>
    <td>Resolved address of the <code>spec.subscriber</code>.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>replyUri</code></td>
    <td>URL (string)</td>
    <td>Resolved address of the <code>spec.reply</code>.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</code></td>
    <td>URL (string)</td>
    <td>Resolved address of the <code>spec.delivery.deadLetterSink</code>.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## TriggerFilter

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>attributes</code></td>
    <td>map[string]string</td>
    <td>Event filter using exact match on event context attributes. Each key in the map is compared with the equivalent key in the event context. An event passes the filter if the event attribute values are equal to the specified values, or if the value in the map is the empty string (which is treated as matching all values but requiring the key to be present).</td>
    <td>REQUIRED</td>
  </tr>
</table>
