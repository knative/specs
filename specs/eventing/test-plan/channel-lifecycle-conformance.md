# Channel Lifecycle 

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#channel-lifecycle


From the Spec: 

>> A Channel represents an Addressable endpoint (i.e. it MUST have a status.address.url field) which can receive, store, and forward events to multiple recipients (Subscriptions).

>> Subscriptions MUST be associated with a Channel based on the spec.channel field on the Subscription; it is expected that the controller for a Channel will also control the associated Subscriptions.

>> When the Channel's Ready condition is true, the Channel MUST provide a status.address.url which accepts all valid CloudEvents and MUST attempt to forward the received events to each associated Subscription whose Ready condition is true. As described in the Subscription Lifecycle section, a Channel MAY forward events to an associated Subscription which does not currently have a true Ready condition, including events received by the Channel before the Subscription was created.

>> When a Channel is created, its spec.channelTemplate field MAY be populated to indicate which of several possible Channel implementations to use. It is RECOMMENDED to default the spec.channelTemplate field on creation if it is unpopulated. Once created, the spec.channelTemplate field MUST be immutable; the Channel MUST be deleted and re-created to change the spec.channelTemplate. This pattern is chosen to make it clear that changing spec.channelTemplate is not an atomic operation and that any implementation would be likely to result in event loss during the transition.



# Testing Channel Lifecycle Conformance: 

We are going to be testing the previous paragraphs coming from the Knative Eventing Spec. To do this we will be creating a Channel, checking its immutable properties, checking its Ready status and then creating a Subscription that links to it by making a reference. We will also checking the Subscription status, as it depends on the Channel to be ready to work correctly. We will be also checking that the Channel is addressable by looking at the status conditions fields. Because this is a Control Plane test, we are not going to be sending Events to these components. 

You can find the resources for running these tests inside the [control-plane/broker-lifecycle/](specs/eventing/test-plan/control-plane/channel-lifecycle/) directory. 
- A [Channel resource](specs/eventing/test-plan/control-plane/channellifecycle/channel.yaml)
- A [Subscription resource that references the Channel](specs/eventing/test-plan/control-plane/channel-lifecycle/trigger.yaml)
- A [Subscription resource that doesn't reference the Channel](specs/eventing/test-plan/control-plane/channel-lifecycle/subscription-no-channel.yaml)


## [Pre] Creating a Channel 

```
kubectl apply -f control-plane/channel-lifecycle/channel.yaml
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


## [Test] Broker Readyness 

Check for condition type `Ready` with status `True`: 

```
 kubectl get broker conformance-broker -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/broker-readyness"
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

## [Test] Trigger for Broker Readyness

Check for condition type `Ready` with status `True`: 

```
kubectl get trigger conformance-trigger -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

### [Output]

```
{
  "test": "control-plane/broker-lifecycle/trigger-for-broker-readyness"
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