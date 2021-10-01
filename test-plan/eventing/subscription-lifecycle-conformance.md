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
- A [Services resource that creates the DeadLetterSink and the Subscriber for the Subscription](control-plane/subscription-lifecycle/services.yaml)


## [Pre] Creating a Subscription 

```
kubectl apply -f control-plane/subscription-lifecycle/subscription.yaml
```

## [Test] Inmutability

Check for default annotations, this should return the name of the selected implementation: 

```
kubectl get subscription conformance-subscription -o jsonpath='{.spec.channel}'
```

Try to patch the annotation: `spec.channel` to see if the resource mutates: 

```
kubectl patch subscription conformance-subscription --type merge -p '{"spec":{"channel":{"apiVersion":"mutable"}}}'
```

You should get the following error: 
```
Error from server (BadRequest): admission webhook "validation.webhook.eventing.knative.dev" denied the request: validation failed: Immutable fields changed (-old +new): spec
{v1.SubscriptionSpec}.Channel.APIVersion:
	-: "messaging.knative.dev/v1"
	+: "mutable"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/subscription-lifecycle/immutability-1"
  "output": {
	  "expectedError": "<EXPECTED ERROR>"
  }
}
```

## [Test] Subscription Readiness 

Check for condition type `Ready` with status `False` cause the some of the preconditions on the spec are not met: 

```
 kubectl get subscription conformance-subscription -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/subscription-lifecycle/subscription-readiness"
  "output": {
    "expectedType": "Ready",
    "expectedStatus": "False"
  }
}
```

## [Pre] Creating the Channel

```
kubectl apply -f control-plane/subscription-lifecycle/channel.yaml
```

## [Test] Subscription Readiness 2

Check for condition type `Ready` with status `False` cause with even with the Channel some of the preconditions on the spec are not met: 

```
 kubectl get subscription conformance-subscription -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Lets now check the reason why the Subscription is not ready yet:

```
 kubectl get subscription conformance-subscription -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].reason}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/subscription-lifecycle/subscription-readiness-2"
  "output": {
	  "expectedType": "Ready",
	  "expectedStatus": "False",
    "expectedReason: "SubscriberResolveFailed"
  }
}
```

## [Pre] Creating the DeadLetterSink and the Subscriber

```
kubectl apply -f control-plane/subscription-lifecycle/services.yaml
```

## [Test] Subscription status update

The `status.physicalSubscription` should have been updated with the 

```
kubectl get subscription conformance-subscription -ojsonpath="{.status.physicalSubscription}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/subscription-lifecycle/subscription-status-update"
  "output": {
	  "obtainedPhysicalSubscription": {
      "deadLetterSinkUri": "<DEADLETTERSINK_URI>",
      "subscriberUri": "<SUBSCRIBER_URI>"
    }
  }
}
```

## [Test] Subscription Readiness 2

Check for condition type `Ready` with status `False` cause with even with the Channel some of the preconditions on the spec are not met: 

```
 kubectl get subscription conformance-subscription -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Lets now check the reason why the Subscription is not ready yet:

```
 kubectl get subscription conformance-subscription -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].reason}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/subscription-lifecycle/subscription-readiness-2"
  "output": {
	  "expectedType": "Ready",
	  "expectedStatus": "False",
    "expectedReason: "SubscriberResolveFailed"
  }
}
```

# Clean up & Congratulations

Make sure that you clean up all resources created in these tests by running: 

```
kubectl delete -f control-plane/subscription-lifecycle/
```


Congratulations you have tested the **Subscription Lifecycle Conformance** :metal: !
