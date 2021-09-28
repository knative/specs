# Channel Lifecycle 

From: https://github.com/knative/specs/blob/release-0.26/specs/eventing/control-plane.md#channel-lifecycle


From the Spec: 

>> A Channel represents an Addressable endpoint (i.e. it MUST have a status.address.url field) which can receive, store, and forward events to multiple recipients (Subscriptions).

>> Subscriptions MUST be associated with a Channel based on the spec.channel field on the Subscription; it is expected that the controller for a Channel will also control the associated Subscriptions.

>> When the Channel's Ready condition is true, the Channel MUST provide a status.address.url which accepts all valid CloudEvents and MUST attempt to forward the received events to each associated Subscription whose Ready condition is true. As described in the Subscription Lifecycle section, a Channel MAY forward events to an associated Subscription which does not currently have a true Ready condition, including events received by the Channel before the Subscription was created.

>> When a Channel is created, its spec.channelTemplate field MAY be populated to indicate which of several possible Channel implementations to use. It is RECOMMENDED to default the spec.channelTemplate field on creation if it is unpopulated. Once created, the spec.channelTemplate field MUST be immutable; the Channel MUST be deleted and re-created to change the spec.channelTemplate. This pattern is chosen to make it clear that changing spec.channelTemplate is not an atomic operation and that any implementation would be likely to result in event loss during the transition.



# Testing Channel Lifecycle Conformance: 

We are going to be testing the previous paragraphs coming from the Knative Eventing Spec. To do this we will be creating a Channel, checking its immutable properties, checking its Ready status and then creating a Subscription that links to it by making a reference. We will also checking the Subscription status, as it depends on the Channel to be ready to work correctly. We will be also checking that the Channel is addressable by looking at the status conditions fields. Because this is a Control Plane test, we are not going to be sending Events to the components. 

You can find the resources for running these tests inside the [control-plane/channel-lifecycle/](test-plan/eventing/channel/control-plane/) directory. 
- A [Channel resource](test-plan/eventing/channel/control-plane/channel.yaml)
- A [Subscription resource that references the Channel](test-plan/eventing/channel/control-plane/subscription.yaml)
- A [Service resource that serves as deadletter sink and subscriber for the subscritpion](test-plan/eventing/channel/control-plane/service.yaml)


## [Pre] Creating a Channel 

```
kubectl apply -f channel/control-plane/channel.yaml
```


## [Test] Immutability

Check for default annotations, this should return the name of the selected implementation: 

```
kubectl get channel.messaging.knative.dev conformance-channel -o jsonpath='{.spec.channelTemplate.kind}'
```

Try to patch the annotation: `messaging.knative.dev/channel.spec.channelTemplate` to see if the resource mutates: 

```
kubectl patch channel.messaging.knative.dev conformance-channel --type merge -p '{"spec":{"channelTemplate":{"kind":"mutable"}}}'
```

You should get the following error: 
```
Error from server (BadRequest): admission webhook "validation.webhook.eventing.knative.dev" denied the request: validation failed: Immutable fields changed (-old +new): annotations
{string}:
	-: "InMemoryChannel" // or your channel implementation
	+: "mutable"
```

Tested in eventing:
- Not tested (depends on the implementation)

### [Output]

```
{
  "test": "channel/control-plane/immutability-1"
  "output": {
    	"channel Implementation": "<CHANNEL IMPLEMENTATION>",
	"expectedError": "<EXPECTED ERROR>"
  }
}
```

## [Test] Channel Readiness 

Check for condition type `Ready` with status `True`: 

```
 kubectl get channel.messaging.knative.dev conformance-channel -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/channel/control_plane.go#L102
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/channel_status_test.go#L29

### [Output]

```
{
  "test": "channel/control-plane/channel-readiness"
  "output": {
  	"channelImplementation": "<CHANNEL IMPLEMENTATION>",
	"expectedType": "Ready",
	"expectedStatus": "True"
  }
}
```

## [Test] Channel is Addressable

Running the following command should return a URL

```
kubectl get channel.messaging.knative.dev conformance-channel -ojsonpath="{.status.address.url}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/channel/control_plane.go#L102
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/channel_status_test.go#L29

### [Output]

```
{
  "test": "channel/control-plane/channel-addressable"
  "output": {
  	"channelImplementation": "",
	"obtainedURL": "<CHANNEL URL>",
  }
}
```

## [Pre] Create Subscription for the Channel

First lets create a Service that works as a Subscriber and a deadLetterSink for the Subscription:

```
kubectl apply -f channel/control-plane/services.yaml
```

Create a Subscription that points to the Channel:

```
kubectl apply -f channel/control-plane/subscription.yaml
```

## [Test] Channel Reference in Subscription

Check that the Subscription is making a reference to the Channel, this should return the name of the Channel.

```
kubectl get subscription conformance-subscription -ojsonpath="{.spec.channel.name}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/resources/subscription/subscription_test.go#L166

### [Output]

```
{
  "test": "channel/control-plane/channel-reference-in-subscription"
  "output": {
  	"channelImplementation": "<CHANNEL IMPLEMENTATION>",
	"expectedReference": "conformance-channel"
  }
}
```

## [Test] Subscription for Channel Readiness

Check for condition type `Ready` with status `True`: 

```
kubectl get subscription conformance-subscription -ojsonpath="{.status.conditions[?(@.type == \"Ready\")].status}"
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/channel/control_plane.go#L84
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/channel_status_subscriber_test.go#L29

### [Output]

```
{
  "test": "channel/control-plane/subscription-for-channel-readiness"
  "output": {
  	"channelImplementation": "<CHANNEL IMPLEMENTATION>",
	"expectedType": "Ready",
	"expectedStatus": "True"
  }
}
```

### [Test] Channel fan out message to Subscribers

Test for messages sent from the channel to each Subscription Subscriber:

Lets create first some Ping Sources to start sending events to the conformance-channel:


```
kubectl apply -f channel/control-plane/ping-sources.yaml
```

Now lets look for those events in each Subscription Subscriber ref logs:

```
kubectl logs --ignore-errors --tail 100 -l serving.knative.dev/service=conformance-events -c user-container | grep conformance-pingsource-1 | tail -n 5

kubectl logs --ignore-errors --tail 100 -l serving.knative.dev/service=conformance-events -c user-container | grep conformance-pingsource-2 | tail -n 5

kubectl logs --ignore-errors --tail 100 -l serving.knative.dev/service=conformance-events -c user-container | grep conformance-pingsource-3 | tail -n 5
```

Tested in eventing:
- https://github.com/knative/eventing/blob/release-0.26/test/rekt/features/channel/topology_test.go#L29
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/channel_tracing_test.go#L30
- https://github.com/knative/eventing/blob/release-0.26/test/conformance/channel_data_plane_test.go#L29

### [Output]

```
{
  "test": "channel/control-plane/channel-fan-out-messages-to-subscribers"
  "output": { 
    *Logs of the messages sent to the different subscription subscribers*
  }
}
```
# Clean up & Congratulations

Make sure that you clean up all resources created by these tests by running: 

```
kubectl delete -f channel/control-plane/
```


Congratulations you have tested the **Channel Lifecycle Conformance** :metal: !