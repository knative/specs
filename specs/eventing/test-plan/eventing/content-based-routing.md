# Content Based Routing

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#content-based-routing

From the Spec:

>> A Broker MUST publish a URL at status.address.url when it is able to receive
>> events. This URL MUST implement the receiver requirements of event delivery.
>> Before acknowledging an event, the Broker MUST durably enqueue the event
>> (where durability means that the Broker can retry event delivery beyond the
>> duration of receiving the event).

>> For each event received by the Broker, the Broker MUST evaluate each
>> associated Trigger exactly once (where "associated" means a Trigger with a
>> spec.broker which references the Broker). If the Trigger has a Ready
>> condition of true when the event is evaluated, the Broker MUST evaluate the
>> Trigger's spec.filter and, if matched, proceed with event delivery as
>> described below. The Broker MAY also evaluate and forward events to
>> associated Triggers for which the Ready condition is not currently true.
>> (One example: a Trigger which is in the process of being programmed in the
>> Broker data plane might receive some events before the data plane
>> programming was complete and the Trigger was updated to set the Ready
>> condition to true.)

>> If multiple Triggers match an event, one event delivery MUST be generated
>> for each match; duplicate matches with the same destination MUST each
>> generate separate event delivery attempts, one per Trigger match. The
>> implementation MAY attach additional event attributes or other metadata
>> distinguishing between these deliveries. The implementation MUST NOT modify
>> the event data in this process.

>> Reply events generated during event delivery MUST be re-enqueued by the
>> Broker using the same routing and persistence as events delivered to the
>> Broker's Addressable URL. Reply events re-enqueued in this manner MUST be
>> evaluated against all Triggers associated with the Broker, including the
>> Trigger that generated the reply. If the storage of the reply event in the
>> Broker fails, the entire event delivery MUST be failed and the delivery to
>> the Trigger's subscriber MUST be retried. Implementations MAY implement
>> event-loop detection; it is RECOMMENDED that any such controls be documented
>> to end-users. Implementations MAY avoid using HTTP to deliver event replies
>> to the Broker's event-delivery input and instead use an internal queueing
>> mechanism.

# Testing Content Based Routing Conformance:

We are going to be testing that, under normal operation (no pod failures), when
an event is sent to a broker, each associated Trigger receives the event exactly
once when it passes the trigger filter. We are going to test different filters.

We are going to send events to the event-display service to manually check that
events are being received.

You can find the resources for running these tests inside the
[control-plane/content-based-routing](control-plane/content-based-routing)
directory.

## [Pre] Creating a Broker

```sh
kubectl apply -f control-plane/content-based-routing/broker.yaml
```

## [Pre] Broker readiness

Check for condition type Ready with status True:

```sh
kubectl get broker conformance-broker -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

### [output]

```sh
{
  "test": "control-plane/content-based-routing/event-delivery-to-broker/pre-1"
  "output": {
    "expectedType": "Ready",
    "expectedStatus": "True"
  }
}
```

## [Pre] Creating an Event Display service

```sh
kubectl apply -f control-plane/content-based-routing/event-display.yaml
```

## [Pre] Event Display service readiness

```sh
kubectl get endpoints event-display -ojsonpath="{.subsets[0].addresses[0].ip}"
```

### [output]:

```sh
{
  "test": "control-plane/content-based-routing/event-delivery-to-broker/pre-2"
  "output": {
    "expectedType": "IP",
    "expectedStatus": "<service endpoint IP>"
  }
}
```

## [Pre] Creating a Trigger with no filters

```sh
kubectl apply -f control-plane/content-based-routing/trigger-no-filter.yaml
```

## [Test] Trigger readiness

```sh
kubectl get triggers.eventing.knative.dev conformance-trigger-no-filter -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

### [output]:

```
{
  "test": "control-plane/content-based-routing/event-delivery-to-broker/test-1"
  "output": {
    "expectedType": "Ready",
    "expectedStatus": "True"
  }
}
```

## [Test] Send an event to the Broker

Sends an event using `kn event`

```sh
kn event send --type dev.knative.conformance.ping --id 1 --to Broker:eventing.knative.dev/v1:conformance-broker
```

### [output]:

```
{
  "test": "control-plane/content-based-routing/event-delivery-to-broker/test-2"
  "output": {
    "receivedCloudEvent": {
      <JSON structure of the CloudEvent received here
    },
  }
}
```

