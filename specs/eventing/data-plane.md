# Knative Eventing Data Plane Contract

## Introduction

Late-binding event senders and receivers (composing applications using
configuration) only works when all event senders and recipients speak a common
protocol. In order to enable wide support for senders and receivers, Knative
Eventing extends the [CloudEvents HTTP
bindings](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md)
with additional semantics for the following reasons:

- Knative Eventing aims to enable highly-reliable event processing workflows. As
  such, it prefers duplicate delivery to discarded events. The CloudEvents spec
  does not take a stance here.

- The CloudEvents HTTP bindings provide a relatively simple and efficient
  network protocol which can easily be supported in a wide variety of
  programming languages leveraging existing library investments in HTTP.

- Knative Eventing assumes a sender-driven (push) event delivery system. That
  is, each event processor is actively responsible for an event until it is
  handled (or affirmatively delivered to all following recipients).
  
- Knative Eventing aims to make writing [event
  sources](./overview.md#event-source) and event-processing software easier to
  write; as such, it imposes higher standards on system components like
  [brokers](./overview.md#broker) and [channels](./overview.md#channel) than on
  edge components.

This contract defines a mechanism for a single event sender to reliably deliver
a single event to a single recipient. Building from this primitive, chains of
reliable event delivery and event-driven applications can be built.

## Background

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in RFC2119.

When not specified in this document, the [CloudEvents HTTP bindings, version
1](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md) and
[HTTP 1.1 protocol](https://tools.ietf.org/html/rfc7230) standards should be
followed (with the CloudEvents bindings preferred in the case of conflict).

The current version of this document does not describe protocol negotiation or
the ability to upgrade an HTTP 1.1 event delivery into a more efficient protocol
such as GRPC, AMQP or the like. It also aims not to preclude such a protocol
negotiation in future versions of the specification.


## Event Delivery

To provide simpler support for event sources which may be translating events
from existing systems, some data plane requirements for senders are relaxed in
the general case. In the case of Knative Eventing provided resources (Channels
and Brokers) which implement these roles, requirements may be increased from
SHOULD to MUST. These cases are called out as they occur.

### Minimum supported protocol

All senders and recipients MUST support the CloudEvents 1.0 protocol and the
[binary](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#31-binary-content-mode)
and
[structured](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#32-structured-content-mode)
content modes of the CloudEvetns HTTP binding. Senders MUST support both
cleartext (`http`) and TLS (`https`) URLs as event delivery destinations.

### HTTP Verbs

In the absence of specific delivery preferences, the sender MUST initiate
delivery of the event to the recipient using the HTTP POST verb, using either
the structured or binary encoding of the event (sender's choice). This delivery
SHOULD be performed using the CloudEvents HTTP Binding, version 1.

Senders MAY probe the recipient with an [HTTP OPTIONS
request](https://tools.ietf.org/html/rfc7231#section-4.3.7); if implemented, the
recipent MUST indicate support the POST verb using the [`Allow`
header](https://tools.ietf.org/html/rfc7231#section-7.4.1). Senders which
receive an error SHOULD proceed using the HTTP POST mechanism.

### Event Acknowledgement and Repeat Delivery

Event recipients MUST use the HTTP response code to indicate acceptance of an
event. The recipient MUST NOT return a response accepting the event until it has
handled event (processed the event or stored in stable storage). The following
response codes are explicitly defined; event recipients MAY also respond with
other response codes. A response code not in this table SHOULD be treated as an
error.

| Response code | Meaning                     | Retry |
| ------------- | --------------------------- | ----- |
| 200           | [Event Reply](#event-reply) | No    |
| 202           | Event accepted              | No    |
| 400           | Unparsable event            | No    |
| 404           | Endpoint does not exist     | No    |

Recipients MUST be able to handle duplicate delivery of events and MUST accept
delivery of duplicate events, as the event acknowledgement could have been lost in
transit to the sender.


### Observability


### Derived (Reply) Events




================================================================
    CUT BELOW
================================================================


## Data plane contract for Sinks

A **Sink** MUST be able to handle duplicate events.

A **Sink** is an [_addressable_](./interfaces.md#addressable) resource that
takes responsibility for the event. A Sink could be a consumer of events, or
middleware. A Sink MUST be able to receive CloudEvents over HTTP and HTTPS.

A **Sink** MAY be [_callable_](./interfaces.md#callable) resource that
represents an Addressable endpoint which receives an event as input and
optionally returns an event to forward downstream.

Almost every component in Knative Eventing may be a Sink providing
composability.

Every Sink MUST support HTTP Protocol Binding for CloudEvents
[version 1.0](https://github.com/cloudevents/spec/blob/v1.0/http-protocol-binding.md)
and
[version 0.3](https://github.com/cloudevents/spec/blob/v0.3/http-transport-binding.md)
with restrictions and extensions specified below.

### HTTP Support

This section adds restrictions on
[requirements in HTTP Protocol Binding for CloudEvents](https://github.com/cloudevents/spec/blob/v1.0/http-protocol-binding.md#12-relation-to-http).

Sinks MUST accept HTTP requests with POST method and MAY support other HTTP
methods. If a method is not supported Sink MUST respond with HTTP status code
`405 Method Not Supported`. Non-event requests (e.g. health checks) are not
constrained.

The URL used by a Sink MUST correspond to a single, unique endpoint at any given
moment in time. This MAY be done via the host, path, query string, or any
combination of these. This mapping is handled exclusively by the
[Addressable control-plane](./interfaces.md#control-plane) exposed via the
`status.address.url`.

If an HTTP request's URL does not correspond to an existing endpoint, then the
Sink MUST respond with `404 Not Found`.

Every non-Callable Sink MUST respond with `202 Accepted` if the request is
accepted.

If Sink is Callable it MAY respond with `200 OK` and a single event in the HTTP
response. A returned event is not required to be related to the received event.
The Callable should return a successful response if the event was processed
successfully. If there is no event to send back then Callable Sink MUST respond
with 2xx HTTP and with empty body.

If a Sink receives a request and is unable to parse a valid CloudEvent, then it
MUST respond with `400 Bad Request`.

### Content Modes Supported

A Sink MUST support `Binary Content Mode` and `Structured Content Mode` as
described in
[HTTP Message Mapping section of HTTP Protocol Binding for CloudEvents](https://github.com/cloudevents/spec/blob/master/http-protocol-binding.md#3-http-message-mapping)

A Sink MAY support `Batched Content Mode` but that mode is not used in Knative
Eventing currently (that may change in future).

### Retries

Sinks should expect that retries and accept possibility that duplicate events
may be delivered.

### Error handling

If Sink is not returning HTTP success header (200 or 202) then the event may be
sent again. If the event can not be delivered then some sources of events (such
as Knative sources, brokers or channels) MAY support
[dead letter sink or channel](https://github.com/knative/eventing/blob/main/docs/delivery/README.md) for events that can not be
delivered.

### Observability

CloudEvents received by Sink MAY have
[Distributed Tracing Extension Attribute](https://github.com/cloudevents/spec/blob/v1.0/extensions/distributed-tracing.md).

### Event reply contract

An event sender supporting event replies SHOULD include a `Prefer: reply` header
in delivery requests to indicate to the sink that event reply is supported. An
event sender MAY ignore an event reply in the delivery response if the
`Prefer: reply` header was not included in the delivery request.

An example is that a Broker supporting event reply sends events with an
additional header `Prefer: reply` so that the sink connected to the Broker knows
event replies will be accepted. While a source sends events without the header,
in which case the sink may assume that any event reply will be dropped without
error or retry attempt. If a sink wishes to ensure the reply events will be
delivered, it can check for the existence of the `Prefer: reply` header in the
delivery request and respond with an error code if the header is not present.

### Data plane contract for Sources

See [Source Delivery specification](sources.md#source-event-delivery)
for details.

### Data plane contract for Channels

See [Channel Delivery specification](channel.md#data-plane) for details.

### Data plane contract for Brokers

See [Broker Delivery specification](broker.md)

## Changelog

- 2020-04-20: `0.13.x release`: initial version that documents common contract
  for sinks, sources, channels and brokers.
