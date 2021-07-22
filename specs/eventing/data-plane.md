# Knative Eventing Data Plane Contract

## Terminology

This document discusses communication between two parties:

- **Event Senders** initiate an HTTP POST to deliver a CloudEvent.
- **Event Recipients** receive an HTTP POST and accept (or reject) a CloudEvent.

Additionally, these roles can be combined in different ways:

- **Event Processors** can be event senders, event recipients, or both.
- **Event Sources** are exclusively event senders, and never act as recipients.
- **Event Sinks** are exclusively event recipients, and do not send events as
  part of their event handling.

## Introduction

Late-binding event senders and recipients (composing applications using
configuration) only works when all event senders and recipients speak a common
protocol. In order to enable wide support for senders and recipients, Knative
Eventing extends the
[CloudEvents HTTP bindings](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md)
with additional semantics for the following reasons:

- Knative Eventing aims to enable at least once event processing; hence it
  prefers duplicate delivery to discarded events. The CloudEvents spec does not
  take a stance here.

- The CloudEvents HTTP bindings provide a relatively simple and efficient
  network protocol which can easily be supported in a wide variety of
  programming languages leveraging existing library investments in HTTP. The
  CloudEvents project has already written these libraries for many popular
  languages.

- Knative Eventing assumes a sender-driven (push) event delivery system. That
  is, each recipient is actively responsible for an event until it is handled
  (or affirmatively delivered to all following recipients).

- Knative Eventing aims to make [event sources](./overview.md#event-source) and
  event-processing software easier to write; as such, it imposes higher
  standards on system components like [brokers](./overview.md#broker) and
  [channels](./overview.md#channel) than on edge components.

This contract defines a mechanism for a single event sender to reliably deliver
a single event to a single recipient. Building from this primitive, chains of
reliable event delivery and event-driven applications can be built.

## Background

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be
interpreted as described in
[RFC2119](https://datatracker.ietf.org/doc/html/rfc2119).

When not specified in this document, the
[CloudEvents HTTP bindings, version 1.0](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md)
and [HTTP 1.1 protocol](https://tools.ietf.org/html/rfc7230) standards MUST be
followed (with the CloudEvents bindings taking precedence in the case of
conflict).

The current version of this document does not describe protocol negotiation or
any delivery mechanism other than HTTP 1.1. Future versions might define
protocol negotiation to optimize delivery; compliant implementations SHOULD aim
to interoperate by ignoring unrecognized negotiation options (such as
[HTTP `Upgrade` headers](https://datatracker.ietf.org/doc/html/rfc7230#section-6.7)).

## Event Delivery

### Minimum supported protocol

All senders and recipients MUST support the CloudEvents 1.0 protocol and the
[binary](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#31-binary-content-mode)
and
[structured](https://github.com/cloudevents/spec/blob/v1.0.1/http-protocol-binding.md#32-structured-content-mode)
content modes of the CloudEvents HTTP binding. Senders which do not advertise
the ability to accept [reply events](#derived-reply-events) MAY implement only
one content mode, as the recipient is not allowed to negotiate the content mode.

### HTTP Verbs

In the absence of specific delivery preferences, the sender MUST initiate
delivery of the event to the recipient using the HTTP POST verb, using either
the structured or binary encoding of the event (sender's choice). This delivery
MUST be performed using the CloudEvents HTTP Binding, version 1.x.

Senders MAY probe the recipient with an
[HTTP OPTIONS request](https://tools.ietf.org/html/rfc7231#section-4.3.7); if
implemented, the recipient MUST indicate support for the POST verb using the
[`Allow` header](https://tools.ietf.org/html/rfc7231#section-7.4.1). Senders
which receive an error when probing with HTTP OPTIONS SHOULD proceed using the
HTTP POST mechanism.

### Event Acknowledgement and Delivery Retry

Event recipients MUST use the HTTP response code to indicate acceptance of an
event. The recipient SHOULD NOT return a response accepting the event until it
has handled the event (processed the event or stored it in stable storage). The
following response codes are explicitly defined; event recipients MAY also
respond with other response codes. A response code not in this table SHOULD be
treated as a retriable error.

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

\* Unspecified `1xx`, `2xx`, and `3xx` response codes are **reserved for future
extension**. Event recipients SHOULD NOT send these response codes in this spec
version, but event senders MUST handle these response codes as errors or success
as appropriate and implement described success or failure behavior.

Recipients MUST accept duplicate delivery of events, but they are NOT REQUIRED
to detect that they are duplicates. If duplicate detection is implemented, then
as specified in the
[CloudEvents specification](https://github.com/cloudevents/spec/blob/v1.0.1/primer.md#id),
event recipients MUST use the
[`source` and `id` attributes](https://github.com/cloudevents/spec/blob/v1.0.1/spec.md#required-attributes)
to identify duplicate events. This specification does not describe state
requirements for recipients which need to detect duplicate events. In general,
senders MAY add or update other CloudEvent attributes on each delivery attempt;
see [observability](#observability) for an example case.

Where possible, event senders SHOULD re-attempt delivery of events where the
HTTP request returned a retryable status code. It is RECOMMENDED that event
senders implement some form of congestion control (such as exponential backoff)
and delivery throttling when managing retry timing. Congestion control MAY cause
event delivery to fail or MAY include not retrying failed delivery attempts.
This specification does not document any specific congestion control algorithm
or parameters. [Brokers](./overview.md#broker) and
[Channels](./overview.md#channel) MUST implement congestion control and MUST
implement retries.

### Observability

Event senders MAY add or update CloudEvents attributes before sending to
implement observability features such as tracing; in particular, the
`traceparent` and `tracestate` distributed tracing attributes defined by
[W3C](https://www.w3.org/TR/trace-context/) and
[CloudEvents](https://github.com/cloudevents/spec/blob/v1.0/extensions/distributed-tracing.md)
MAY be modified in this way for each delivery attempt of the same event.

This specification does not mandate any particular logging or metrics
aggregation, nor a method of exposing observability information to users
configuring the resources. Platform administrators SHOULD expose event-delivery
telemetry to users through platform-specific interfaces, but such interfaces are
beyond the scope of this document.

### Derived (Reply) Events

In some applications, an event recipient MAY emit an event in reaction to a
received event. Senders MAY choose to support this pattern by accepting an
encoded CloudEvent in the HTTP response.

An event sender MAY document support for this pattern by including a
`Prefer: reply` header in the HTTP POST request. This header indicates to the
event recipient that the caller will accept a
[`200` response](#event-acknowledgement-and-repeat-delivery) which includes a
CloudEvent encoded using the binary or structured formats.
[Brokers](./overview.md#broker) and [Channels](./overview.md#channel) MUST
indicate support for replies using the `Prefer: reply` header when sending to
the `spec.subscriber` address.

A recipient MAY reply to any HTTP POST with a `200` response to indicate that
the event was processed successfully, with or without a response payload. If the
recipient does not produce a response payload, the `202` response code is also
acceptable. Responses with a `202` response code MUST NOT be processed as reply
events; even if the response can be interpreted as a CloudEvent, the status
monitor for the accepted-but-not-completed request MUST NOT be routed further.

If a recipient chooses to reply to a sender with a `200` response code and a
reply event in the absence of a `Prefer: reply` header from the sender, the
sender SHOULD treat the event as accepted, and MAY log an error about the
unexpected payload. If a sender will process a reply event it MUST include the
`Prefer: reply` header on the POST request.
