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
such as GRPC, AMQP or the like. It is expected that a future compatible version
of this specification might describe a protocol negotiation mechanism.

## Event Delivery

To provide simpler support for event sources which might be translating events
from existing systems, some data plane requirements for senders are relaxed in
the general case. In the case of Knative Eventing provided resources (Channels
and Brokers) which implement these roles, requirements are increased from
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
SHOULD be performed using the CloudEvents HTTP Binding, version 1.0.

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
other response codes. A response code not in this table SHOULD be treated as a
retriable error.

| Response code | Meaning                     | Retry | Delivery completed | Error |
| ------------- | --------------------------- | ----- | ------------------ | ----- |
| `1xx`         | (Unspecified)               | Yes\* | No\*               | Yes\* |
| `200`         | [Event reply](#event-reply) | No    | Yes                | No    |
| `202`         | Event accepted              | No    | Yes                | No    |
| other `2xx`   | (Unspecified)               | Yes\* | No\*               | Yes\* |
| other `3xx`   | (Unspecified)               | Yes\* | No\*               | Yes\* |
| `400`         | Unparsable event            | No    | No                 | Yes   |
| `404`         | Endpoint does not exist     | Yes   | No                 | Yes   |
| other `4xx`   | Error                       | Yes   | No                 | Yes   |
| other `5xx`   | Error                       | Yes   | No                 | Yes   |

\* `1xx`, `2xx`, and `3xx` response codes are **reserved for future
extension**. Event recipients SHOULD NOT send these response codes in this spec
version, but event senders MUST handle these response codes as errors and
implement appropriate failure behavior.

<!-- TODO: Should 3xx redirects and 401 (Unauthorized) or 403 (Forbidden) errors
be retried? What about `405` (Method Not Allowed), 413 (Payload Too Large), 414
(URI Too Long), 426 (Upgrade Required), 431 (Header Fields Too Large), 451
(Unavailable for Legal Reasons)? -->

Recipients MUST be able to handle duplicate delivery of events and MUST accept
delivery of duplicate events, as the event acknowledgement could have been lost
in transit to the sender. Event recipients MUST use the [`source` and `id`
attributes](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md#required-attributes)
to detect duplicated events (see [observability](#observability) for an example
case where other event attributes may vary from one delivery attempt to
another).

Where possible, event senders SHOULD re-attempt delivery of events where the
HTTP request failed or returned a retriable status code. It is RECOMMENDED that
event senders implement some form of congestion control (such as exponential
backoff) when managing retry timing. This specification does not document any
specific congestion control algorithm or
parameters. [Brokers](./overview.md#broker) and
[Channels](./overview.md#channel) MUST implement congestion control and MUST
implement retries.

### Observability

Event senders MAY add or update CloudEvents attributes before sending to
implement observability features such as tracing; in particular, the
[`traceparent` and `tracestate` distributed tracing
attributes](https://github.com/cloudevents/spec/blob/v1.0/extensions/distributed-tracing.md)
may be modified in this way for each delivery attempt of the same event.

This specification does not mandate any particular logging or metrics
aggregtion, nor a method of exposing observability information to users
configuring the resources. Platform administrators SHOULD expose event-delivery
telemetry to users through platform-specific interfaces, but such interfaces are
beyond the scope of this document.

<!-- TODO: should we mentioned RECOMMENDED spans or RECOMMENDED metrics like in
https://github.com/knative/specs/blob/main/specs/eventing/channel.md#observability?
-->

### Derived (Reply) Events

In some applications, an event receiver might emit an event in reaction to a
received event. An event sender MAY document support for this pattern by
including a `Prefer: reply` header in the HTTP POST request. This header
indicates to the event receiver that the caller will accept a [`200`
response](#event-acknowledgement-and-repeat-delivery) which includes a
CloudEvent encoded using the binary or structured formats.

The sender SHOULD NOT assume that a received reply event is directly related to
the event sent in the HTTP request.

A recipient may reply to any HTTP POST with a `200` response to indicate that
the event was processed successfully, with or without a response payload. If the
recipient will _never_ provide a response payload, the `202` response code is
likely a better choice.

If a recipient chooses to reply to a sender with a `200` response code and a
reply event in the absence of a `Prefer: reply` header, the sender SHOULD treat
the event as accepted, and MAY log an error about the unexpected payload. The
sender MUST NOT process the reply event if it did not advertise the `Prefer:
reply` capability.
