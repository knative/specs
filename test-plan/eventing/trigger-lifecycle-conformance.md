# Broker Lifecycle 

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#trigger-lifecycle


From the Spec: 

>> A Trigger MAY be created before the referenced Broker indicated by its spec.broker field; if the Broker does not currently exist or the Broker's Ready condition is not true, then the Trigger's Ready condition MUST be false, and the reason SHOULD indicate that the corresponding Broker is missing or not ready.

>> The Trigger's controller MUST also set the status.subscriberUri field based on resolving the spec.subscriber field before setting the Ready condition to true. If the spec.subscriber.ref field points to a resource which does not exist or cannot be resolved via Destination resolution, the Trigger MUST set the Ready condition to false, and at least one condition MUST indicate the reason for the error. The Trigger MUST also set status.subscriberUri to the empty string if the spec.subscriber.ref cannot be resolved.

>> If the Trigger's spec.delivery.deadLetterSink field it set, it MUST be resolved to a URI and reported in status.deadLetterSinkUri in the same manner as the spec.subscriber field before setting the Ready condition to true.

>> Once created, the Trigger's spec.broker MUST NOT permit updates; to change the spec.broker, the Trigger can instead be deleted and re-created. This pattern is chosen to make it clear that changing the spec.broker is not an atomic operation, as it could span multiple storage systems. Changes to spec.subscriber, spec.filter and other fields SHOULD be permitted, as these could occur within a single storage system.

>> When a Trigger becomes associated with a Broker (either due to creating the Trigger or the Broker), the Trigger MUST only set the Ready condition to true after the Broker has been configured to send all future events matching the spec.filter to the Trigger's spec.subscriber. The Broker MAY send some events to the Trigger's spec.subscriber prior to the Trigger's Readycondition being set to true. When a Trigger is deleted, the Broker MAY send some additional events to the Trigger's spec.subscriber after the deletion.

# Testing Trigger Lifecycle Conformance: 

We are going to be testing the previous paragraphs coming from the Knative Eventing Spec. To do this we will be creating a broker, checking its immutable properties, checking its Ready status and then creating a Broker that reference to it by making a reference. Because this is a Control Plane test, we are not going to be sending Events to these components. 

You can find the resources for running these tests inside the [control-plane/trigger-lifecycle/](control-plane/broker-lifecycle/) directory. 
- A [Trigger resource](control-plane/trigger-lifecycle/trigger.yaml)
- A [Broker resource that references the Broker](trigger-lifecycle/broker.yaml)
- A [Trigger resource that doesn't reference the Broker](control-plane/trigger-lifecycle/trigger-no-broker.yaml)


## [Pre] Creating a Broker 

```
kubectl apply -f control-plane/broker-lifecycle/broker.yaml
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
