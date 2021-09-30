# Trigger Lifecycle 

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#trigger-lifecycle


From the Spec: 

>> A Trigger MAY be created before the referenced Broker indicated by its spec.broker field; if the Broker does not currently exist or the Broker's Ready condition is not true, then the Trigger's Ready condition MUST be false, and the reason SHOULD indicate that the corresponding Broker is missing or not ready.

>> The Trigger's controller MUST also set the status.subscriberUri field based on resolving the spec.subscriber field before setting the Ready condition to true. If the spec.subscriber.ref field points to a resource which does not exist or cannot be resolved via Destination resolution, the Trigger MUST set the Ready condition to false, and at least one condition MUST indicate the reason for the error. The Trigger MUST also set status.subscriberUri to the empty string if the spec.subscriber.ref cannot be resolved.

>> If the Trigger's spec.delivery.deadLetterSink field it set, it MUST be resolved to a URI and reported in status.deadLetterSinkUri in the same manner as the spec.subscriber field before setting the Ready condition to true.

>> Once created, the Trigger's spec.broker MUST NOT permit updates; to change the spec.broker, the Trigger can instead be deleted and re-created. This pattern is chosen to make it clear that changing the spec.broker is not an atomic operation, as it could span multiple storage systems. Changes to spec.subscriber, spec.filter and other fields SHOULD be permitted, as these could occur within a single storage system.

>> When a Trigger becomes associated with a Broker (either due to creating the Trigger or the Broker), the Trigger MUST only set the Ready condition to true after the Broker has been configured to send all future events matching the spec.filter to the Trigger's spec.subscriber. The Broker MAY send some events to the Trigger's spec.subscriber prior to the Trigger's Readycondition being set to true. When a Trigger is deleted, the Broker MAY send some additional events to the Trigger's spec.subscriber after the deletion.

# Testing Trigger Lifecycle Conformance: 

We are going to be testing the previous paragraphs coming from the Knative Eventing Spec. To do this we will be creating a Trigger, checking its immutable properties, checking its Ready status and then creating a Broker that is referenced by it. Because this is a Control Plane test, we are not going to be sending Events to these components. 

You can find the resources for running these tests inside the [control-plane/trigger-lifecycle/](control-plane/broker-lifecycle/) directory. 
- A [Trigger resource](control-plane/trigger-lifecycle/1-trigger.yaml)
- A [Broker resource that is referenced by the Trigger](trigger-lifecycle/broker.yaml)
- A [Trigger resource that have a non resolvable Subscriber URI](control-plane/trigger-lifecycle/trigger-no-subscriber.yaml)


## [Pre] Creating a Trigger

Lets create a Trigger that does not have a valid reference to a Broker yet:

```
kubectl apply -f control-plane/trigger-lifecycle/1-trigger.yaml
```


## [Test] Immutability

Check for the Broker reference in the spec, this must be inmmutable: 

```
kubectl get trigger conformance-trigger -o jsonpath='{.spec.broker}'
```

Try to patch the spec Broker reference: `spec.broker` to see if the resource mutates: 

```
kubectl patch trigger conformance-trigger --type merge -p '{"spec":{"broker":"mutable"}}'
```

You should get the following error: 
```
Error from server (BadRequest): admission webhook "validation.webhook.eventing.knative.dev" denied the request: validation failed: Immutable fields changed (-old +new): annotations
{string}:
	-: "conformance-broker"
	+: "mutable"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/trigger-lifecycle/immutability-1"
  "output": {
    "brokerReference": "<CONFORMANCE_BROKER>",
	  "expectedError": "<EXPECTED ERROR>"
  }
}
```

## [Test] Trigger Readiness 

Check for condition type `Ready` with status `False` since there is no Broker related to the Trigger: 

```
 kubectl get trigger conformance-trigger -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/trigger-lifecycle/trigger-readiness"
  "output": {
    "expectedType": "Ready",
    "expectedStatus": "False"
  }
}
```

## [Test] Trigger Subscriber is resolvable

Running the following command should return a URI:

```
kubectl get trigger conformance-trigger -ojsonpath="{.spec.subscriber}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/trigger-lifecycle/trigger-subscriber-resolvable"
  "output": {
	"obtainedURI": "<SUBSCRIBER_URI>",
  }
}
```

## [Test] Trigger Sink is resolvable

Running the following command should return a URI:

```
kubectl get trigger conformance-trigger -ojsonpath="{.spec.subscriber}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/trigger-lifecycle/trigger-sink-resolvable"
  "output": {
	"obtainedURI": "<SUBSCRIBER_URI>",
  }
}
```

## [Pre] Create Broker for Trigger

Create a Broker to be referenced by the Trigger:

```
kubectl apply -f control-plane/trigger-lifecycle/broker.yaml
```

## [Test] Test for Trigger Readiness

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
  "test": "control-plane/trigger-lifecycle/trigger-for-broker-readiness"
  "output": {
	"expectedType": "Ready",
	"expectedStatus": "True"
  }
}
```


## [Pre] Create Trigger with a non resolvable Subscriber URI

Create a Trigger that have a non resolvable Subscriber URI:

```
kubectl apply -f control-plane/trigger-lifecycle/trigger-no-subscriber.yaml
```

## [Test] Trigger subscriber is not resolvable

Running the following command should return a URL

```
kubectl get trigger conformance-trigger-no-subscriber -ojsonpath="{.spec.subscriber}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/trigger-lifecycle/trigger-subscriber-not-resolvable"
  "output": {
	"obtainedURI": "<SUBSCRIBER_REF>",
  }
}
```

## [Test] Trigger readdiness when subscriber is not resolvable

Check for condition type `Ready` with status `False` since there is no Subscriber resolvable URI related to the Trigger: 

```
 kubectl get trigger conformance-trigger-no-subscriber -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Now lets check if there is a clear reason indicating what is wrong"

```
 kubectl get trigger conformance-trigger-no-subscriber -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].reason}"
```

Finally lets see if the `status.subscriberUri` is empty:

```
kubectl get trigger conformance-trigger-no-subscriber -ojsonpath="{.status.subscriberUri}"
```

Tested in eventing:
- 

### [Output]

```
{
  "test": "control-plane/trigger-lifecycle/trigger-readiness-subscriber-not-resolvable"
  "output": {
    "expectedType": "Ready",
    "expectedStatus": "False"
    "expectedReason: "Unable to get the Subscriber's URI"
    "expectedUri": ""
  }
}
```

# Clean up & Congratulations

Make sure that you clean up all resources created in these tests by running: 

```
kubectl delete -f control-plane/trigger-lifecycle/
```


Congratulations you have tested the **Trigger Lifecycle Conformance** :metal: !
