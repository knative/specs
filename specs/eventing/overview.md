# Resource Types

The Knative Eventing API provides primitives for two common event-processing
patterns (credit to James Urquhart for the formulation):

* Point-to-point asynchronous communication ([`messaging.knative.dev`](#messaging))

* Content-based event routing ([`eventing.knative.dev`](#eventing))

The other two patterns James identifies are log-stream processing and complex
workflows; these are not currently addressed by Knative Eventing.

In addition to the primitives needed to express the above patterns, Knative
Eventing defines two [_interface contracts_](#interface-contracts) to allow
connecting multiple types of Kubernetes objects as event senders and recipients
to the core primitives.

<!-- TODO: add a drawing -->

## Eventing

### Broker

**Broker** provides a central event-routing hub which exposes a URL address
which event senders may use to submit events to the router. A Broker may be
implemented using many different underlying event-forwarding mechanisms; the
broker provides a small set of common event-delivery configuration options and
may reference additional implementation-specific configuration options via a
reference to an external object; the format of the external objects is not
standardized.

### Trigger

**Trigger** defines a filtered delivery option to extract events delivered to a
**Broker** and route them to an **Addressable** destination. Trigger implements
uniform event filtering based on the CloudEvents attributes associated with the
event, ignoring the payload (which might be large and/or binary and need not be
parsed during event routing). The addressable interface contract allows Triggers
to deliver events to a variety of different destinations, including external
resources such as a virtual machine or SaaS service.

## Messaging

### Channel

**Channel** provides an abstract interface which may be fulfiled by several
concrete implementations of a backing asynchronous fan-out queue. The common
abstraction provided by channel allows both the composition of higher-level
constructs for chained or parallel processing of events, and the replacement of
particular messaging technologies (for example, allowing a development
environment to use a lower-reliability channel compared with the production
environment).

### Subscription

**Subscription** defines a delivery destination for all events sent to a
**Channel**. Events sent to a channel are delivered to _each_ subscription
_independently_ -- a subscription maintains its own list of undelivered events
and will manage retry indpendently of any other subscriptions to the same
channel. Like **Trigger**, subscriptions use the **Addressable** interface
contract to support event delivery to many different destination types.

## Interface Contracts

In addition to the concrete types described above in the `messaging.knative.dev`
and `eventing.knative.dev` API groups, Knative Eventing supports referencing
objects in other API groups as destinations for event delivery. This is done by
defining partial schemas which the other resources must support. The following
interface contracts define a set of expected resource fields on an referenced
resource.

### Addressable

**Addressable** resources expose a resource address (HTTP URL) in their `status`
object. The URL is used as a destination for delivery of events to the resource;
the exposed URL must implement the [data plane contract](data-plane.md) for
receiving events.

**Broker** and **Channel** both implement **Addressable**.

### Event Source

**Event Sources** are resources which generate events and may be configured to
deliver the events to an **Addressable** resource designated by a `sink` object
in the resource's `spec`. The Knative Eventing spec does not define any specific
event sources, but does define common interfaces for discovering and managing
event sources.

