# Subscription Lifecycle 

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#subscription-lifecycle


From the Spec: 

>> A Subscription MAY be created before the referenced Channel indicated by its spec.channel field. The spec.channel object reference MAY refer to either a messaging.knative.dev/v1 Channel resource, or another resource which meets the spec.subscribers and spec.delivery required elements in the Channelable duck type. The spec.channel reference MUST be to an object in the same namespace; specifically, the spec.channel.namespace field must be unset or the empty string. If the referenced spec.channel does not currently exist or its Ready condition is not true, then the Subscription's Ready condition MUST NOT be true, and the reason SHOULD indicate that the corresponding Channel is missing or not ready. 

>> The Subscription MUST also set the status.physicalSubscription URIs by resolving the spec.subscriber, spec.reply, and spec.delivery.deadLetterSink as described in Destination resolution before setting the Ready condition to true. If any of the addressable fields fails resolution, the Subscription MUST set the Ready condition to false, and at least one condition MUST indicate the reason for the error. The Subscription MUST also set status.physicalSubscription URIs to the empty string if the corresponding spec reference cannot be resolved.

>> The spec.subscriber destination MUST be set; if the spec.reply field is not set, replies from the spec.subscriber MUST be discarded.

>> Once created, the Subscription's spec.channel MUST NOT permit updates; to change the spec.channel, the Subscription MUST be deleted and re-created. This pattern is chosen to make it clear that changing the spec.channel is not an atomic operation, as it might span multiple storage systems. Changes to spec.subscriber, spec.reply, spec.delivery and other fields SHOULD be permitted, as these could occur within a single storage system.

>> When a Subscription becomes associated with a Channel (either due to creating the Subscription or the Channel), the Subscription MUST only set the Ready condition to true after the Channel has been configured to send all future events to the Subscription's spec.subscriber. The Channel MAY send some events to the Subscription prior to the Subscription's Ready condition being set to true. When a Subscription is deleted, the Channel MAY send some additional events to the Subscription's spec.subscriber after the deletion.

# Testing Subscription Lifecycle Conformance: 

We are going to be testing the previous paragraphs coming from the Knative Eventing Spec. To do this we will be creating a Subscription, checking its immutable properties, checking its Ready status and then creating a Channel that is referenced by it. We will also checking the Subscription status, as it depends on the Channel/Ref to be ready to work correctly. We will be also checking that the status is updated with its corresponding URIs. Because this is a Control Plane test, we are not going to be sending Events to these components. 

You can find the resources for running these tests inside the [control-plane/subscription-lifecycle/](control-plane/subscription-lifecycle/) directory. 
- A [Subscription resource](control-plane/subscription-lifecycle/subscription.yaml)
- A [Channel resource that is referenced by the Subscription](subscription-lifecycle/channel.yaml)
- A [Subscription resource that doesn't reference the Channel](control-plane/subscription-lifecycle/subscription-no-channel.yaml)


## [Pre] Creating a Subscription 

```
kubectl apply -f control-plane/subscription-lifecycle/subscription.yaml
```


## [Test] Immutability

Check for default annotations, this should return the name of the selected implementation: 

```
kubectl get broker conformance-broker -o jsonpath='{.metadata.annotations.eventing\.knative\.dev/broker\.class}'
```

Try to patch the annotation: `eventing.knative.dev/broker.class` to see if the resource mutates: 

```
kubectl patch broker conformance-broker --type merge -p '{"metadata":{"annotations":{"eventing.knative.dev/broker.class":"mutable"}}}'
```

You should get the following error: 
```
Error from server (BadRequest): admission webhook "validation.webhook.eventing.knative.dev" denied the request: validation failed: Immutable fields changed (-old +new): annotations
{string}:
	-: "MTChannelBasedBroker"
	+: "mutable"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/broker/control_plane.go#L90

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/immutability-1"
  "output": {
    	"brokerImplementation": "<BROKER IMPLEMENTATION>",
	"expectedError": "<EXPECTED ERROR>"
  }
}
```

Try to mutate the `.spec.config` to see if the resource mutates: 

```
kubectl patch broker conformance-broker --type merge -p '{"spec":{"config":{"apiVersion":"v1"}}}'
```


### [Output]

```
{
  "test": "control-plane/broker-lifecycle/immutability-2"
  "output": {
  	"brokerImplementation": "<BROKER IMPLEMENTATION>",
	"expectedError": "<EXPECTED ERROR>"
  }
}
```


## [Test] Broker Readiness 

Check for condition type `Ready` with status `True`: 

```
 kubectl get broker conformance-broker -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/helpers/broker_control_plane_test_helper.go#L104
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/broker/control_plane.go#L86

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/broker-readiness"
  "output": {
  	"brokerImplementation": "<BROKER IMPLEMENTATION>",
	"expectedType": "Ready",
	"expectedStatus": "True"
  }
}
```

## [Test] Broker is Addresable

Running the following command should return a URL

```
kubectl get broker conformance-broker -ojsonpath="{.status.address.url}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/helpers/broker_control_plane_test_helper.go#L109
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/broker/control_plane.go#L88

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/broker-addressable"
  "output": {
  	"brokerImplementation": "",
	"obtainedURL": "<BROKER URL>",
  }
}
```

## [Pre] Create Trigger for Broker

Create a trigger that points to the broker:

```
kubectl apply -f control-plane/broker-lifecycle/trigger.yaml
```

## [Test] Broker Reference in Trigger

Check that the `Trigger` is making a reference to the `Broker`, this should return the name of the broker.

```
kubectl get trigger conformance-trigger -ojsonpath="{.spec.broker}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/broker/control_plane.go#L114

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/broker-reference-in-trigger"
  "output": {
  	"brokerImplementation": "<BROKER IMPLEMENTATION>",
	"expectedReference": "conformance-broker"
  }
}
```

## [Test] Trigger for Broker Readiness

Check for condition type `Ready` with status `True`: 

```
kubectl get trigger conformance-trigger -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/helpers/broker_control_plane_test_helper.go#L139
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/broker/control_plane.go#L112

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/trigger-for-broker-readiness"
  "output": {
  	"brokerImplementation": "<BROKER IMPLEMENTATION>",
	"expectedType": "Ready",
	"expectedStatus": "True"
  }
}
```

# Clean up & Congratulations

Make sure that you clean up all resources created in these tests by running: 

```
kubectl delete -f control-plane/broker-lifecycle/
```


Congratulations you have tested the **Broker Lifecycle Conformance** :metal: !
