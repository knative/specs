# Knative Control Plane Contract

## Abstract

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

## Background

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be
interpreted as described in [RFC 2119](https://tools.ietf.org/html/rfc2119).

There is no formal specification of the Kubernetes API and Resource Model. This
document assumes Kubernetes 1.21 behavior; this behavior will typically be
supported by many future Kubernetes versions. Additionally, this document
references specific core Kubernetes resources; these references can be
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

## Resource Overview

The Knative Eventing API provides a set of primitives to support both
point-to-point communication channels (`messaging.knative.dev`) and
content-based event routing (`eventing.knative.dev`). This specification
describes API interfaces of Knative Eventing resources as well as the supported
[event routing](./data-plane.md) logic and configuration settings.

At the moment, the Knative Eventing specification does not contemplate any
non-Kubernetes-backed implementations, and therefore does not specifically
define the mapping of Kubernetes verbs (read, watch, patch, etc) to developer
roles. See the [Overview documentation](./overview.md) for general definitions
of the different API objects.

## RBAC Profile

In order to validate the controls described in
[Resource Overview](#resource-overview), the following Kubernetes RBAC profile
can be applied in a Kubernetes cluster. This Kubernetes RBAC is an illustrative
example of the minimal profile rather than a requirement. This Role is
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
which wish to participate in [Addressable resolution](#addressable-resolution) are expected to provide the
following `ClusterRole`:

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

This configuration advice does not indicate a requirement for Kubernetes RBAC
support.

<!-- TODO: define aggregated API roles

Ref:
- https://github.com/knative/specs/blob/main/specs/eventing/channel.md#aggregated-channelable-manipulator-clusterrole
- https://github.com/knative/specs/blob/main/specs/eventing/channel.md#aggregated-addressable-resolver-clusterrole
-
-->

## Error Signalling

See [the Knative common condition guidance](../common/error-signalling.md) for
how resource errors are signalled to the user.

## Resource Lifecycle

### Broker Lifecycle

A Broker represents an Addressable endpoint (i.e. it MUST have a
`status.address.url` field) which can receive, store, and forward events to
multiple recipients based on a set of attribute filters (Triggers). Triggers
MUST be associated with a Broker based on the `spec.broker` field on the
Trigger; it is expected that the controller for a Broker will also control the
associated Triggers. When the Broker's `Ready` condition is `true`, the Broker
MUST provide a `status.address.url` which accepts all valid CloudEvents and MUST
attempt to forward the received events for filtering to each associated Trigger
whose `Ready` condition is `true`. As described in the
[Trigger Lifecycle](#trigger-lifecycle) section, a Broker MAY forward events to
an associated Trigger which which does not currently have a `true` `Ready`
condition, including events received by the Broker before the Trigger was
created.

The annotation `eventing.knative.dev/broker.class` MAY be used to select a
particular implementation of a Broker. When a Broker is created, the
`eventing.knative.dev/broker.class` annotation and the `spec.config` field MUST
be populated (`spec.config` MAY be an empty object) to indicate which of several
possible Broker implementations to use. It is RECOMMENDED to default the
`eventing.knative.dev/broker.class` field on creation if it is unpopulated. Once
created, both fields MUST be immutable; the Broker MUST be deleted and
re-created to change the implementation class or `spec.config`. This pattern is
chosen to make it clear that changing the implementation class or `spec.config`
is not an atomic operation and that any implementation would be likely to result
in event loss during the transition.

### Trigger Lifecycle

A Trigger MAY be created before the referenced Broker indicated by its
`spec.broker` field; if the Broker does not currently exist or the Broker's
`Ready` condition is not `true`, then the Trigger's `Ready` condition MUST be
`false`, and the reason SHOULD indicate that the corresponding Broker is missing
or not ready.

The Trigger's controller MUST also set the `status.subscriberUri` field based on
resolving the `spec.subscriber` field before setting the `Ready` condition to
`true`. If the `spec.subscriber.ref` field points to a resource which does not
exist or cannot be resolved via
[Addressable resolution](#addressable-resolution), the Trigger MUST set the
`Ready` condition to `false`, and at least one condition MUST indicate the
reason for the error. The Trigger SHOULD also set `status.subscriberUri` to the
empty string if the `spec.subscriber.ref` cannot be resolved.

If the Trigger's `spec.delivery.deadLetterSink` field it set, it MUST be
resolved to a URI and reported in `status.deadLetterSinkUri` in the same manner
as the `spec.subscriber` field before setting the `Ready` condition to `true`.

Once created, the Trigger's `spec.broker` MUST NOT permit updates; to change the
`spec.broker`, the Trigger can instead be deleted and re-created. This pattern
is chosen to make it clear that changing the `spec.broker` is not an atomic
operation, as it could span multiple storage systems. Changes to
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

### Channel Lifecycle

A Channel represents an Addressable endpoint (i.e. it MUST have a
`status.address.url` field) which can receive, store, and forward events to
multiple recipients (Subscriptions). Subscriptions MUST be associated with a
Channel based on the `spec.channel` field on the Subscription; it is expected
that the controller for a Channel will also control the associated
Subscriptions. When the Channel's `Ready` condition is `true`, the Channel MUST
provide a `status.address.url` which accepts all valid CloudEvents and MUST
attempt to forward the received events to each associated Subscription whose
`Ready` condition is `true`. As described in the
[Subscription Lifecycle](#subscription-lifecycle) section, a Channel MAY forward
events to an associated Subscription which does not currently have a `true`
`Ready` condition, including events received by the Channel before the
`Subscription` was created.

When a Channel is created, its `spec.channelTemplate` field MUST be populated to
indicate which of several possible Channel implementations to use. It is
RECOMMENDED to default the `spec.channelTemplate` field on creation if it is
unpopulated. Once created, the `spec.channelTemplate` field MUST be immutable;
the Channel MUST be deleted and re-created to change the `spec.channelTemplate`.
This pattern is chosen to make it clear that changing `spec.channelTemplate` is
not an atomic operation and that any implementation would be likely to result in
event loss during the transition.

### Subscription Lifecycle

A Subscription MAY be created before the referenced Channel indicated by its
`spec.channel` field. The `spec.channel` object reference MAY refer to either an
`messaging.knative.dev/v1` Channel resource, or another resource which meets the
`spec.subscribers` and `spec.delivery` required elements in the Channelable duck
type. If the referenced `spec.channel` does not currently exist or its `Ready`
condition is not `true`, then the Subscription's `Ready` condition MUST NOT be
`true`, and the reason SHOULD indicate that the corresponding Channel is missing
or not ready.

The Subscription MUST also set the `status.physicalSubscription` URIs by
resolving the `spec.subscriber`, `spec.reply`, and
`spec.delivery.deadLetterSink` as described in
[Addressable resolution](#addressable-resolution) before setting the `Ready`
condition to `true`. If any of the addressable fields fails resolution, the
Subscription MUST set the `Ready` condition to `false`, and at least one
condition MUST indicate the reason for the error. The Subscription SHOULD also
set `status.physicalSubscription` URIs to the empty string if the corresponding
`spec` reference cannot be resolved.

At least one of `spec.subscriber` and `spec.reply` MUST be set; if only
`spec.reply` is set, the behavior is equivalent to setting `spec.subscriber`
except that the Channel SHOULD NOT advertise the ability to process replies
during the delivery.

Once created, the Subscription's `spec.channel` MUST NOT permit updates; to
change the `spec.channel`, the Subscription can be deleted and re-created. This
pattern is chosen to make it clear that changing the `spec.channel` is not an
atomic operation, as it might span multiple storage systems. Changes to
`spec.subscriber`, `spec.reply`, `spec.delivery` and other fields SHOULD be
permitted, as these could occur within a single storage system.

When a Subscription becomes associated with a Channel (either due to creating
the Subscription or the Channel), the Subscription MUST only set the `Ready`
condition to `true` after the Channel has been configured to send all future
events to the Subscriptions `spec.subscriber`. The Channel MAY send some events
to the Subscription before prior to the Subscription's `Ready` condition being
set to `true`. When a Subscription is deleted, the Channel MAY send some
additional events to the Subscription's `spec.subscriber` after the deletion.

<!--
TODO: channel-compatible CRDs (Channelable)
-->

### Addressable Resolution

Both Trigger and Subscription have OPTIONAL object references (`ref` in
`spec.subscriber`, `spec.delivery.deadLetterSink`, and `spec.reply` for
Subscription) which are expected to conform to the Addressable partial schema
("duck type"). An object conforms with the Addressable partial schema if it
contains a `status.address.url` field containing a URL which can be used to
deliver CloudEvents over HTTP. As a special case, Kubernetes `v1` `Service`
objects are considered to have a `status.address.url` of
`http://<service-dns-name>/`.

If both the `ref` field and the `uri` fields are set on a Destination, then the
Destination's address MUST be the `uri` interpreted relative to the resolved URI
of the `ref` field. This can be used, for example, to refer to a specific URL
off a referenced domain name, like so:

```yaml
subscriber:
  ref:
    apiVersion: v1
    kind: Service
    name: test
  uri: "/update"
```

This Destination would resolve to
`http://test.<current-namespace-dns-prefix>/update`, as `/update` would be
interpreted relative to the `test` Service's DNS name.

If one of these object references points to an object which does not currently
satisfy this partial schema (because object does not exist, the
`status.address.url` field is empty, or because the object does not have that
field), then the Trigger or Subscription MUST indicate an error by setting the
`Ready` condition to `false`, and SHOULD include an indication of the error in a
condition reason or type.

Both Broker and Channel MUST conform to the Addressable partial schema.

## Event Routing

Note that the event routing description below does not cover the actual
mechanics of sending an event from one component to another; see
[the data plane](./data-plane.md) contracts for details of the event transfer
mechanisms.

### Content Based Routing

A Broker MUST publish a URL at `status.address.url` when it is able to receive
events. This URL MUST accept CloudEvents in both the
[Binary Content Mode](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#31-binary-content-mode)
and
[Structured Content Mode](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#32-structured-content-mode)
HTTP formats. Before sending an HTTP response, the Broker MUST durably enqueue
the event (where durability means that the Broker can retry event delivery
beyond the duration of receiving the event).

For each event received by the Broker, the Broker MUST evaluate each associated
Trigger **exactly once** (where "associated" means a Trigger with a
`spec.broker` which references the Broker). If the Trigger has a `Ready`
condition of `true` when the event is evaluated, the Broker MUST evaluate the
Trigger's `spec.filter` and, if matched, proceed with
[event delivery as described below](#event-delivery). The Broker MAY also
evaluate and forward events to associated Triggers for which the `Ready`
condition is not currently `true`. (One example: a Trigger which is in the
process of being programmed in the Broker data plane might receive _some_ events
before the data plane programming was complete and the Trigger was updated to
set the `Ready` condition to `true`.)

If multiple Triggers match an event, one event delivery MUST be generated for
each match; duplicate matches with the same destination MUST each generate a
separate event delivery attempts, one per Trigger match. The implementation MAY
attach additional event attributes or other metadata distinguishing between
these deliveries. The implementation MUST NOT modify the event payload in this
process.

Reply events generated during event delivery MUST be re-enqueued by the Broker
in the same way as events delivered to the Broker's Addressable URL. If the
storage of the reply event fails, the entire event delivery MUST be failed and
the delivery to the Trigger's subscriber MUST be retried. Reply events
re-enqueued in this manner MUST be evaluated against all triggers associated
with the Broker, including the Trigger that generated the reply. Implementations
MAY implement event-loop detection; it is RECOMMENDED that any such controls be
documented to end-users. Implementations MAY avoid using HTTP to deliver event
replies to the Broker's event-delivery input and instead use an internal
queueing mechanism.

### Topology Based Routing

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

### Event Delivery

Once a Trigger or Subscription has decided to deliver an event, it MUST do the
following:

1. Resolve all URLs and delivery options, using the values in `status` for URL
   resolution.

1. Attempt delivery to the `status.subscriberUri` URL following the
   [data plane contract](./data-plane.md).

   1. If the event delivery fails with a retryable error, it MUST be retried up
      to `retry` times (subject to congestion control), following the
      `backoffPolicy` and `backoffDelay` parameters if specified.

1. If the delivery attempt is successful (either the original request or a
   retry) and no event is returned, the event delivery is complete.

1. If the delivery attempt is successful (either the original request or a
   retry) and an event is returned in the reply, the reply event will be
   delivered to the `status.replyUri` destination (for Subscriptions) or added
   to the Broker for processing (for Triggers). If `status.replyUri` is not
   present in the Subscription, the reply event MUST be dropped.

   1. For Subscriptions, if delivery of the reply event fails with a retryable
      error, the entire delivery of the event to MUST be retried up to `retry`
      times (subject to congestion control), following the `backoffPolicy` and
      `backoffDelay` parameters if specified.

1. If all retries are exhausted for either the original delivery or the retry,
   or if a non-retryable error is received, the event MUST be delivered to the
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

## Detailed Resources

The following schema defines a set of REQUIRED or RECOMMENDED resource fields on
the Knative resource types. Whether a field is REQUIRED or RECOMMENDED is
denoted in the "Schema Requirement" column. Additional `spec` and `status`
fields MAY be provided by particular implementations, however it is expected
that most extension will be accomplished via the `metadata.labels` and
`metadata.annotations` fields, as Knative implementations MAY validate supplied
resources against these fields and refuse resources which specify unknown
fields. Knative implementations MUST NOT require `spec` fields outside this
implementation; to do so would break interoperability between such
implementations and implementations which implement validation of field names.

### Broker

#### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource. The `apiVersion` is `eventing.knative.dev/v1` and the `kind` is
`Broker`.

#### Spec:

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>config</code></td>
    <td><a href="#kreference">KReference</a><br/>(OPTIONAL)</td>
    <td>A reference to an object which describes the configuration options for the Broker (for example, a ConfigMap).</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>A default delivery options for Triggers which do not specify more-specific options. If a Trigger specifies _any_ delivery options, this field MUST be ignored.</td>
    <td>REQUIRED</td>
  </tr>
</table>

#### Status

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
    <td>Used for signalling errors, see link.</td>
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
    <td><a href="#duckv1addressable">duckv1.Addressable</a></td>
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

### Trigger

#### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource. The `apiVersion` is `eventing.knative.dev/v1` and the `kind` is
`Trigger`.

#### Spec:

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
    <td>The Broker to which this Trigger is associated.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>filter</code></td>
    <td><a href="#triggerfilter">TriggerFilter</a><br/>(OPTIONAL)</td>
    <td>Event filters which are used to select events to be delivered to the Trigger's destination.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>subscriber</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a><br/>(Required)</td>
    <td>The destination for delivery of filtered events.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>Delivery options for this Trigger.</td>
    <td>RECOMMENDED</td>
  </tr>
</table>

#### Status

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
    <td>Used for signalling errors, see link.</td>
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

### Channel

#### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource. The `apiVersion` is `messaging.knative.dev/v1` and the `kind` is
`Channel`.

#### Spec:

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
    <td>[]<a href="#duckv1subscriberspec">duckv1.SubscriberSpec</a></td>
    <td>Aggregated subscription information; this array MUST be managed automatically by the controller.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>Default delivery options for Subscriptions which do not specify more-specific options. If a Subscription specifies _any_ delivery options, this field MUST be ignored.</td>
    <td>REQUIRED</td>
  </tr>
</table>

#### Status

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
    <td>Used for signalling errors, see link.</td>
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
    <td><a href="#duckv1addressable">duckv1.Addressable</a></td>
    <td>Address used to deliver events to the Broker.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>subscribers</code></td>
    <td>[]<a href="#duckv1subscriberstatus">duckv1.SubscriberStatus</a></td>
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

### Subscription

#### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource. The `apiVersion` is `messaging.knative.dev/v1` and the `kind` is
`Subscription`.

#### Spec:

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
    <td><a href="#duckv1destination">duckv1.Destination</a><br/>(Required)</td>
    <td>The destination for event delivery.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>reply</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a><br/>(Required)</td>
    <td>The destination for reply events from <code>spec.subscriber</code>.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>Delivery options for this Subscription.</td>
    <td>RECOMMENDED</td>
  </tr>
</table>

#### Status

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
    <td>Used for signalling errors, see link.</td>
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
    <td><a href="#physicalsubscriptionstatus">PhysicalSubscriptionStatus</a></td>
    <td>The fully resolved values for <code>spec</code> endpoint references.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### Addressable

Note that the Addressable interface is a partial schema -- any resource which
includes these fields MAY be referenced using a `duckv1.Destination`.

#### Metadata:

Standard Kubernetes
[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta)
resource. Note that there are no restrictions on `apiVersion` or `kind` as long
as the object matches the partial schema.

#### Spec:

There are no `spec` requirements for Addressable.

#### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td><a href="#duckv1addressable">duckv1.Addressable</a></td>
    <td>Address used to deliver events to the resource.</td>
    <td>REQUIRED</td>
  </tr>
</table>

## Detailed SubResource Objects

### duckv1.Addressable

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>url</code></td>
    <td>URL (string)</td>
    <td>Address used to deliver events to the Addressable.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### duckv1.Destination

Destination is used to indicate the destination for event delivery. A
Destination eventually resolves the supplied information to a URL by resolving
`uri` relative to the address of `ref` (if provided) as described in [Addressable resolution](#addressable-resolution). `ref` MAY be an
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
    <td><a href="#duckv1kreference">duckv1.KReference</a><br/>(OPTIONAL)</td>
    <td>An ObjectReference to an Addressable reference to deliver events to.</td>
    <td>REQUIRED</td>
  </tr>
  <tr>
    <td><code>uri</code></td>
    <td>URL (string)<br/>(OPTIONAL)</td>
    <td>A resolved URL to deliver events to.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### duckv1.SubscriberSpec

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
    <td>REQUIRED</td>
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
    <td><a href="#deliveryspec">DeliverySpec</a></td>
    <td>The resolved Subscription delivery options. The <code>deadLetterSink</code> SHOULD use the <code>uri</code> form.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### duckv1.SubscriberStatus

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
    <td>string</td>
    <td>A human readable message indicating details of <code>ready</code> status.</td>
    <td>REQUIRED</td>
  </tr>
</table>

### DeliverySpec

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
    <td><strong>Schema Requirement</strong></td>
  </tr>
  <tr>
    <td><code>deadLetterSink</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a></td>
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
    <td>The initial delay when retrying delivery, in ISO 8601 format.</td>
    <td>RECOMMENDED</td>
  </tr>
  <tr>
    <td><code>backoffPolicy</code></td>
    <td>enum<br/>("linear", "exponential")</td>
    <td>Retry timing scaling policy. Linear policy uses the same <code>backoffDelay</code> for each attempt; Exponential policy uses 2^N multiples of <code>backoffDelay</code></td>
    <td>RECOMMENDED</td>
  </tr>
</table>

### KReference

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

### PhysicalSubscriptionStatus

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

### TriggerFilter

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
    <td>Event filter using exact match on event context attributes. Each key in the map MUST be compared with the equivalent key in the event context. All keys MUST match (as described below) the event attributes for the event to be selected by the Trigger.
    <br>
    For each key specified in the filter, an attribute with that name MUST be present in the event to match. If the value corresponding to the key is non-empty, the value MUST be an exact match to attribute value in the event; an empty string MUST match all attribute values.</td>
    <td>REQUIRED</td>F
  </tr>
</table>
