# Knative Eventing Test Specification

## Introduction

## Control Plane Assertions

### Error Signalling

TODO: See [the Knative common condition guidance](../common/error-signalling.md)
for how resource errors are signalled to the user.

### Resource Lifecycle

#### Broker Lifecycle

A Broker represents an [Addressable endpoint](#destination-resolution) (i.e. it
MUST have a `status.address.url` field) which can receive, store, and forward
events to multiple recipients based on a set of attribute filters (Triggers).

Triggers MUST be associated with a Broker based on the `spec.broker` field on
the Trigger; it is expected that the controller for a Broker will also control
the associated Triggers.

When the Broker's `Ready` condition is `true`, the Broker MUST provide a
`status.address.url` which accepts all valid CloudEvents and MUST attempt to
forward the received events for filtering to each associated Trigger whose
`Ready` condition is `true`.

Once created, the `eventing.knative.dev/broker.class` annotation and the
`spec.config` field MUST be immutable;

the Broker MUST be deleted and re-created to change the implementation class or
`spec.config`.

#### Trigger Lifecycle

The Trigger's controller MUST also set the `status.subscriberUri` field based on
resolving the `spec.subscriber` field before setting the `Ready` condition to
`true`. If the `spec.subscriber.ref` field points to a resource which does not
exist or cannot be resolved via
[Destination resolution](#destination-resolution), the Trigger MUST set the
`Ready` condition to `false`, and at least one condition MUST indicate the
reason for the error.

The Trigger MUST also set `status.subscriberUri` to the empty string if the
`spec.subscriber.ref` cannot be resolved.

If the Trigger's `spec.delivery.deadLetterSink` field it set, it MUST be
resolved to a URI and reported in `status.deadLetterSinkUri` in the same manner
as the `spec.subscriber` field before setting the `Ready` condition to `true`.

Once created, the Trigger's `spec.broker` MUST NOT permit updates.

When a Trigger becomes associated with a Broker (either due to creating the
Trigger or the Broker), the Trigger MUST only set the `Ready` condition to
`true` after the Broker has been configured to send all future events matching
the `spec.filter` to the Trigger's `spec.subscriber`.

#### Channel Lifecycle

A Channel represents an [Addressable endpoint](#destination-resolution) (i.e. it
MUST have a `status.address.url` field) which can receive, store, and forward
events to multiple recipients (Subscriptions).

Subscriptions MUST be associated with a Channel based on the `spec.channel`
field on the Subscription; it is expected that the controller for a Channel will
also control the associated Subscriptions.

When the Channel's `Ready` condition is `true`, the Channel MUST provide a
`status.address.url` which accepts all valid CloudEvents and MUST attempt to
forward the received events to each associated Subscription whose `Ready`
condition is `true`.

Once created, the `spec.channelTemplate` field MUST be immutable; the Channel
MUST be deleted and re-created to change the `spec.channelTemplate`.

### Subscription Lifecycle

The `spec.channel` reference MUST be to an object in the same namespace;
specifically, the `spec.channel.namespace` field must be unset or the empty
string.

If the referenced `spec.channel` does not currently exist or its `Ready`
condition is not `true`, then the Subscription's `Ready` condition MUST NOT be
`true`, and the reason SHOULD indicate that the corresponding Channel is missing
or not ready.

The Subscription MUST also set the `status.physicalSubscription` URIs by
resolving the `spec.subscriber`, `spec.reply`, and
`spec.delivery.deadLetterSink` as described in
[Destination resolution](#destination-resolution) before setting the `Ready`
condition to `true`.

If any of the addressable fields fails resolution, the Subscription MUST set the
`Ready` condition to `false`, and at least one condition MUST indicate the
reason for the error.

The Subscription MUST also set `status.physicalSubscription` URIs to the empty
string if the corresponding `spec` reference cannot be resolved.

At least one of `spec.subscriber` and `spec.reply` MUST be set; if only
`spec.reply` is set, the behavior is equivalent to setting `spec.subscriber`
except that the Channel SHOULD NOT
[advertise the ability to process replies](data-plane.md#derived-reply-events)
during the delivery.

Once created, the Subscription's `spec.channel` MUST NOT permit updates; to
change the `spec.channel`, the Subscription MUST be deleted and re-created.

When a Subscription becomes associated with a Channel (either due to creating
the Subscription or the Channel), the Subscription MUST only set the `Ready`
condition to `true` after the Channel has been configured to send all future
events to the Subscription's `spec.subscriber`.

### Destination Resolution

Destination provides a mechanism to resolve an object reference to an absolute
URL; two mechanisms MUST be supported:

1. Objects conforming to the [Addressable partial schema](#duckv1addressable)
   ("duck type") contain a `status.address.url` field providing a URL which can
   be used to deliver CloudEvents over HTTP.
2. As a special case, Kubernetes `v1/Service` objects are resolved to the
   service's cluster-local DNS name (of the form
   `http://<service-name>.<namespace-name>.svc/`).

If both the `ref` field and the `uri` fields are set on a Destination, then the
Destination's address MUST be the `uri` interpreted relative to the resolved URI
of the `ref` field.

If a Destination includes a reference to an object which does not resolve to an
absolute URL (because object does not exist, the `status.address.url` field is
empty, etc), then the Trigger or Subscription MUST indicate an error by setting
the `Ready` condition to `false`, and SHOULD include an indication of the error
in a condition reason or type.

Both Broker and Channel MUST conform to the
[Addressable partial](#duckv1addressable) schema.

### Event Routing

#### Content Based Routing

A Broker MUST publish a URL at `status.address.url` when it is able to receive
events.

This URL MUST implement the receiver requirements of
[event delivery](#data-plane.md#event-delivery).

Before
[acknowledging an event](data-plane.md#event-acknowledgement-and-delivery-retry),
the Broker MUST durably enqueue the event (where durability means that the
Broker can retry event delivery beyond the duration of receiving the event).

For each event received by the Broker, the Broker MUST evaluate each associated
Trigger **exactly once** (where "associated" means a Trigger with a
`spec.broker` which references the Broker).

If the Trigger has a `Ready` condition of `true` when the event is evaluated,
the Broker MUST evaluate the Trigger's `spec.filter` and, if matched, proceed
with [event delivery as described below](#event-delivery).

The Broker MAY also evaluate and forward events to associated Triggers for which
the `Ready` condition is not currently `true`.

If multiple Triggers match an event, one event delivery MUST be generated for
each match; duplicate matches with the same destination MUST each generate
separate event delivery attempts, one per Trigger match.

The implementation MUST NOT modify the
[event data](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md#event-data)
in this process.

Reply events generated during event delivery MUST be re-enqueued by the Broker
using the same routing and persistence as events delivered to the Broker's
Addressable URL.

Reply events re-enqueued in this manner MUST be evaluated against all Triggers
associated with the Broker, including the Trigger that generated the reply.

If the storage of the reply event in the Broker fails, the entire event delivery
MUST be failed and the delivery to the Trigger's subscriber MUST be retried.

### Topology Based Routing

A Channel MUST publish a URL at `status.address.url` when it is able to receive
events.

This URL MUST implement the receiver requirements of
[event delivery](#data-plane.md#event-delivery).

Before
[acknowledging an event](data-plane.md#event-acknowledgement-and-delivery-retry),
the Channel MUST durably enqueue the event (be able to deliver with retry
without receiving the event again).

For each event received by the Channel, the Channel MUST deliver the event to
each associated Subscription **at least once** (where "associated" means a
Subscription with a `spec.channel` which references the Channel).

If the Subscription has a `Ready` condition of `true` when the event is
evaluated, the Channel MUST forward the event
[as described in event delivery](#event-delivery).

If multiple Subscriptions with the same destination are associated with the same
Channel, each Subscription MUST generate one delivery attempt per Subscription.

The implementation MUST NOT modify the
[event data](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md#event-data)
in this process. (Retry)

### Event Delivery

Once a Trigger or Subscription has decided to deliver an event, it MUST do the
following:

1. Read the resolved URLs and delivery options from the object's `status`
   fields.

1. Attempt delivery to the `status.subscriberUri` URL following the
   [data plane contract](./data-plane.md).

   1. If the event delivery fails with a retryable error, it MUST be retried up
      to `retry` times (subject to congestion control), following the
      `backoffPolicy` and `backoffDelay` parameters if specified.

1. If the delivery attempt is successful (either the original request or a
   retry) and no event is returned, the event delivery is complete.

1. If the delivery attempt is successful (either the original request or a
   retry) and an event is returned in the reply, the reply event MUST be
   delivered to the `status.replyUri` destination (for Subscriptions) or added
   to the Broker for processing (for Triggers). If `status.replyUri` is not
   present in the Subscription, the reply event MUST be dropped.

   1. For Subscriptions, if delivery of the reply event fails with a retryable
      error, the entire delivery of the event to MUST be retried up to `retry`
      times (subject to congestion control), following the `backoffPolicy` and
      `backoffDelay` parameters if specified.

1. If an event (either the initial event or a reply) cannot be delivered, the
   event MUST be delivered to the `deadLetterSink` in the delivery options. If
   no `deadLetterSink` is specified, the event is dropped.

## Detailed Resources

The following schema defines a set of REQUIRED resource fields on the Knative
resource types. All implementations MUST include all schema fields in their API,
though implementations MAY implement validation of fields. Additional `spec` and
`status` fields MAY be provided by particular implementations, however it is
expected that most API extensions will be accomplished via the `metadata.labels`
and `metadata.annotations` fields, as Knative implementations MAY validate
supplied resources against these fields and refuse resources which specify
unknown fields. Knative implementations MUST NOT require `spec` fields outside
this implementation; to do so would break interoperability between such
implementations and implementations which implement validation of field names.

For fields set in a resource `spec`, the "Field Types" column indicates whether
implementations are REQUIRED to validate that a field is set in requests, or
whether the a request is valid if the field is omitted. Field in a resource
`status` MUST be set by the server implementation.

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
  </tr>
  <tr>
    <td><code>config</code></td>
    <td><a href="#kreference">KReference</a><br/>(OPTIONAL)</td>
    <td>A reference to an object which describes the configuration options for the Broker (for example, a ConfigMap).</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>A default delivery options for Triggers which do not specify more-specific options. If a Trigger specifies <strong>any</strong> delivery options, this field MUST be ignored.</td>
  </tr>
</table>

#### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link.</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest <code>metadata.generation</code> that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td><a href="#duckv1addressable">duckv1.Addressable</a></td>
    <td>Address used to deliver events to the Broker.</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</td>
    <td>URL (string)</td>
    <td>If <code>spec.delivery.deadLetterSink</code> is specified, the resolved URL of the dead letter address.</td>
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
  </tr>
  <tr>
    <td><code>broker</code></td>
    <td>string<br/>(REQUIRED, IMMUTABLE)</td>
    <td>The Broker to which this Trigger is associated.</td>
  </tr>
  <tr>
    <td><code>filter</code></td>
    <td><a href="#triggerfilter">TriggerFilter</a><br/>(OPTIONAL)</td>
    <td>Event filters which are used to select events to be delivered to the Trigger's destination.</td>
  </tr>
  <tr>
    <td><code>subscriber</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a><br/>(REQUIRED)</td>
    <td>The destination for delivery of filtered events.</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>Delivery options for this Trigger.</td>
  </tr>
</table>

#### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link.</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest <code>metadata.generation</code> that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
  </tr>
  <tr>
    <td><code>subscriberUri</code></td>
    <td>URL (string)</td>
    <td>The resolved address of the <code>spec.subscriber</code>.</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</td>
    <td>URL (string)</td>
    <td>If <code>spec.delivery.deadLetterSink</code> is specified, the resolved URL of the dead letter address.</td>
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
  </tr>
  <tr>
    <td><code>channelTemplate</code></td>
    <td>object<br/>(OPTIONAL)</td>
    <td>Implementation-specific parameters to configure the channel.</td>
  </tr>
  <tr>
    <td><code>subscribers</code></td>
    <td>[]<a href="#duckv1subscriberspec">duckv1.SubscriberSpec</a> (FILLED BY SERVER)</td>
    <td>Aggregated subscription information; this array MUST be managed automatically by the controller.</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>Default delivery options for Subscriptions which do not specify more-specific options. If a Subscription specifies _any_ delivery options, this field MUST be ignored.</td>
  </tr>
</table>

#### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link.</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest <code>metadata.generation</code> that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
  </tr>
  <tr>
    <td><code>address</code></td>
    <td><a href="#duckv1addressable">duckv1.Addressable</a></td>
    <td>Address used to deliver events to the Broker.</td>
  </tr>
  <tr>
    <td><code>subscribers</code></td>
    <td>[]<a href="#duckv1subscriberstatus">duckv1.SubscriberStatus</a></td>
    <td>Resolved addresses for the <code>spec.subscribers</code> (subscriptions to this Channel).</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</td>
    <td>URL (string)</td>
    <td>If <code>spec.delivery.deadLetterSink</code> is specified, the resolved URL of the dead letter address.</td>
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
  </tr>
  <tr>
    <td><code>channel</code></td>
    <td><a href="#kreference">KReference</a><br/>(REQUIRED, IMMUTABLE)</td>
    <td>The channel this subscription receives events from. <code>namespace</code>  may not be set (must refer to a <a href="#channel">Channel</a> in the same namespace). Immutable.</td>
  </tr>
  <tr>
    <td><code>subscriber</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a><br/>(OPTIONAL)</td>
    <td>The destination for event delivery.</td>
  </tr>
  <tr>
    <td><code>reply</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a><br/>(OPTIONAL)</td>
    <td>The destination for reply events from <code>spec.subscriber</code>.</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a><br/>(OPTIONAL)</td>
    <td>Delivery options for this Subscription.</td>
  </tr>
</table>

#### Status

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>conditions</code></td>
    <td><a href="#error-signalling">See Error Signalling</a></td>
    <td>Used for signalling errors, see link.</td>
  </tr>
  <tr>
    <td><code>observedGeneration</code></td>
    <td>int64</td>
    <td>The latest <code>metadata.generation</code> that the reconciler has attempted. If <code>observedGeneration</code> is updated, <code>conditions</code> MUST be updated with current status in the same transaction.</td>
  </tr>
  <tr>
    <td><code>physicalSubscription</code></td>
    <td><a href="#physicalsubscriptionstatus">PhysicalSubscriptionStatus</a></td>
    <td>The fully resolved values for <code>spec</code> endpoint references.</td>
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
  </tr>
  <tr>
    <td><code>address</code></td>
    <td><a href="#duckv1addressable">duckv1.Addressable</a></td>
    <td>Address used to deliver events to the resource.</td>
  </tr>
</table>

## Detailed SubResource Objects

### duckv1.Addressable

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>url</code></td>
    <td>URL (string)</td>
    <td>Address used to deliver events to the Addressable.</td>
  </tr>
</table>

### duckv1.Destination

Destination is used to indicate the destination for event delivery. A
Destination eventually resolves the supplied information to a URL by resolving
`uri` relative to the address of `ref` (if provided) as described in
[Destination resolution](#destination-resolution).

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>ref</code></td>
    <td><a href="#duckv1kreference">duckv1.KReference</a><br/>(OPTIONAL)</td>
    <td>An ObjectReference to a cluster resource to deliver events to.</td>
  </tr>
  <tr>
    <td><code>uri</code></td>
    <td>URL (string)<br/>(OPTIONAL)</td>
    <td>A URL (possibly relative to <code>ref</code>) to deliver events to.</td>
  </tr>
</table>

### duckv1.SubscriberSpec

SubscriberSpec represents an automatically-populated extraction of information
from a [Subscription](#subscription). SubscriberSpec should be populated by the
server.

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>uid</code></td>
    <td>UID (string)</td>
    <td>UID is used to disambiguate Subscriptions which might be recreated.</td>
  </tr>
  <tr>
    <td><code>generation</code></td>
    <td>int64</td>
    <td>Generation of the copied Subscription.</td>
  </tr>
  <tr>
    <td><code>subscriberUri</code></td>
    <td>URL (string)</td>
    <td>The resolved address of the Subscription's <code>spec.subscriber</code>.</td>
  </tr>
  <tr>
    <td><code>replyUri</code></td>
    <td>URL (string)</td>
    <td>The resolved address of the Subscription's <code>spec.reply</code>.</td>
  </tr>
  <tr>
    <td><code>delivery</code></td>
    <td><a href="#deliveryspec">DeliverySpec</a></td>
    <td>The resolved Subscription delivery options. The <code>deadLetterSink</code> SHOULD use the <code>uri</code> form.</td>
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
  </tr>
  <tr>
    <td><code>uid</code></td>
    <td>UID (string)</td>
    <td>UID is used to disambiguate Subscriptions which might be recreated.</td>
  </tr>
  <tr>
    <td><code>generation</code></td>
    <td>int64</td>
    <td>Generation of the copied Subscription.</td>
  </tr>
  <tr>
    <td><code>ready</code></td>
    <td>kubernetes v1/ConditionStatus</td>
    <td>Ready status of the Subscription's programming into the Channel data plane.</td>
  </tr>
  <tr>
    <td><code>message</code></td>
    <td>string</td>
    <td>A human readable message indicating details of <code>ready</code> status.</td>
  </tr>
</table>

### DeliverySpec

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>deadLetterSink</code></td>
    <td><a href="#duckv1destination">duckv1.Destination</a> (OPTIONAL)</td>
    <td>Fallback address used to deliver events which cannot be delivered during the flow. An implementation MAY place limits on the allowed destinations for the <code>deadLetterSink</code>.</td>
  </tr>
  <tr>
    <td><code>retry</code></td>
    <td>int (OPTIONAL)</td>
    <td>Retry is the minimum number of retries the sender should attempt when sending an event before moving it to the dead letter sink.</td>
  </tr>
  <tr>
    <td><code>backoffDelay</code></td>
    <td>string (OPTIONAL)</td>
    <td>The initial delay when retrying delivery, in ISO 8601 format.</td>
  </tr>
  <tr>
    <td><code>backoffPolicy</code></td>
    <td>enum<br/>["linear", "exponential"] (OPTIONAL)</td>
    <td>Retry timing scaling policy. Linear policy uses the same <code>backoffDelay</code> for each attempt; Exponential policy uses 2^N multiples of <code>backoffDelay</code></td>
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
  </tr>
  <tr>
    <td><code>apiVersion</code></td>
    <td>string (REQUIRED)</td>
    <td>ApiVersion of the target reference.</td>
  </tr>
  <tr>
    <td><code>kind</code></td>
    <td>string (REQUIRED)</td>
    <td>Kind of the target reference.</td>
  </tr>
  <tr>
    <td><code>name</code></td>
    <td>string (REQUIRED)</td>
    <td>Name of the target resource.</td>
  </tr>
  <tr>
    <td><code>namespace</code></td>
    <td>string (OPTIONAL)</td>
    <td>Namespace of the target resource. If unspecified, defaults to the same namespace</td>
  </tr>
</table>

### PhysicalSubscriptionStatus

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>subscriberUri</code></td>
    <td>URL (string)</td>
    <td>Resolved address of the <code>spec.subscriber</code>.</td>
  </tr>
  <tr>
    <td><code>replyUri</code></td>
    <td>URL (string)</td>
    <td>Resolved address of the <code>spec.reply</code>.</td>
  </tr>
  <tr>
    <td><code>deadLetterSinkUri</code></td>
    <td>URL (string)</td>
    <td>Resolved address of the <code>spec.delivery.deadLetterSink</code>.</td>
  </tr>
</table>

### TriggerFilter

<table>
  <tr>
    <td><strong>Field Name</strong></td>
    <td><strong>Field Type</strong></td>
    <td><strong>Description</strong></td>
  </tr>
  <tr>
    <td><code>attributes</code></td>
    <td>map[string]string (OPTIONAL)</td>
    <td>Event filter using exact match on event context attributes. Each key in the map MUST be compared with the equivalent key in the event context. All keys MUST match (as described below) the event attributes for the event to be selected by the Trigger.
    <br>
    For each key specified in the filter, an attribute with that name MUST be present in the event to match. If the value corresponding to the key is non-empty, the value MUST be an exact (case-sensitive) match to attribute value in the event; an empty string MUST match all attribute values.</td>
  </tr>
</table>

## Data Plane Assertions

### Event Delivery

#### Minimum supported protocol

- All senders and recipients MUST support the CloudEvents 1.0 protocol and the
  [binary](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#31-binary-content-mode)
  and
  [structured](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#32-structured-content-mode)
  content modes of the CloudEvents HTTP binding.

#### HTTP Verbs

In the absence of specific delivery preferences, the sender MUST initiate
delivery of the event to the recipient using the HTTP POST verb, using either
the structured or binary encoding of the event (sender's choice).

This delivery MUST be performed using the CloudEvents HTTP Binding, version 1.x.

Senders MAY probe the recipient with an
[HTTP OPTIONS request](https://tools.ietf.org/html/rfc7231#section-4.3.7); if
implemented, the recipient MUST indicate support for the POST verb using the
[`Allow` header](https://tools.ietf.org/html/rfc7231#section-7.4.1).

#### Event Acknowledgement and Delivery Retry

Event recipients MUST use the HTTP response code to indicate acceptance of an
event.

| Response code | Meaning                                           | Retry | Delivery completed | Error |
| ------------- | ------------------------------------------------- | ----- | ------------------ | ----- |
| `1xx`         | (Unspecified)                                     | No\*  | No\*               | Yes\* |
| `200`         | [Accepted, event in reply](#derived-reply-events) | No    | Yes                | No    |
| `202`         | Event accepted                                    | No    | Yes                | No    |
| other `2xx`   | (Unspecified)                                     | No    | Yes                | No    |
| `3xx`         | (Unspecified)                                     | No\*  | No\*               | Yes\* |
| `400`         | Unparsable event                                  | No    | No                 | Yes   |
| `404`         | Endpoint does not exist                           | Yes   | No                 | Yes   |
| `409`         | Conflict / Processing in progress                 | Yes   | No                 | Yes   |
| `429`         | Too Many Requests / Overloaded                    | Yes   | No                 | Yes   |
| other `4xx`   | Error                                             | No    | No                 | Yes   |
| `5xx`         | Error                                             | Yes   | No                 | Yes   |

\* Event recipients SHOULD NOT send these response codes in this spec version,
but event senders MUST handle these response codes as errors or success as
appropriate and implement described success or failure behavior.

Recipients MUST accept duplicate delivery of events, but they are NOT REQUIRED
to detect that they are duplicates.

If duplicate detection is implemented, then as specified in the
[CloudEvents specification](https://github.com/cloudevents/spec/blob/v1.0.1/primer.md#id),
event recipients MUST use the
[`source` and `id` attributes](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md#required-attributes)
to identify duplicate events.

[Brokers](./overview.md#broker) and [Channels](./overview.md#channel) MUST
implement congestion control and MUST implement retries.

#### Observability

#### Derived (Reply) Events

[Brokers](./overview.md#broker) and [Channels](./overview.md#channel) MUST
indicate support for replies using the `Prefer: reply` header when sending to
the `spec.subscriber` address.

Responses with a `202` response code MUST NOT be processed as reply events; even
if the response can be interpreted as a CloudEvent, the status monitor for the
accepted-but-not-completed request MUST NOT be routed further.

If a sender will process a reply event it MUST include the `Prefer: reply`
header on the POST request.
