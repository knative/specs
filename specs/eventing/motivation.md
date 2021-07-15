# Motivation

The goal of the Knative Eventing project is to define common primitives to
enable composing event-processing applications through configuration, rather
than application code.

Building by combining independent components provides a number of benefits for
application designers:

1. Services are loosely coupled during development and may be deployed
   independently on a variety of platforms (Kubernetes, VMs, SaaS or FaaS). This
   composability allows re-use of common patterns and building blocks, even
   across programming language and tooling boundaries.

1. A producer can generate events before a consumer is listening, and a consumer
   can express an interest in an event or class of events that is not yet being
   produced. This allows event-driven applications to evolve over time without
   needing to closely coordinate changes.

1. Services can be connected to create new applications:
   - without modifying producer or consumer
   - with the ability to select a specific subset of events from a particular
     producer

In order to enable loose coupling and late-binding of event producers and
consumers, Knative Eventing utilizes and extends the
[CloudEvents specification](https://github.com/cloudevents/spec) as the data
plane protocol between components. Knative Eventing prioritizes at-least-once
delivery semantics, using the CloudEvents HTTP POST (push) transport as a
minimum common transport between components.

Knative Eventing also defines patterns to simplify the construction and usage of
event producers and consumers.
