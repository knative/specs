# Broker Lifecycle 

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#broker-lifecycle


From the Spec: 

>> A Broker represents an Addressable endpoint (i.e. it MUST have a status.address.url field) which can receive, store, and forward events to multiple recipients based on a set of attribute filters (Triggers). 

>> Triggers MUST be associated with a Broker based on the spec.broker field on the Trigger; it is expected that the controller for a Broker will also control the associated Triggers. 

>> When the Broker's Ready condition is true, the Broker MUST provide a status.address.url which accepts all valid CloudEvents and MUST attempt to forward the received events for filtering to each associated Trigger whose Ready condition is true. As described in the Trigger Lifecycle section, a Broker MAY forward events to an associated Trigger destination which does not currently have a true Ready condition, including events received by the Broker before the Trigger was created.

>> The annotation eventing.knative.dev/broker.class SHOULD be used to select a particular implementation of a Broker, if multiple implementations are available. It is RECOMMENDED to default the eventing.knative.dev/broker.class field on creation if it is unpopulated. Once created, the eventing.knative.dev/broker.class annotation and the spec.config field MUST be immutable; the Broker MUST be deleted and re-created to change the implementation class or spec.config. This pattern is chosen to make it clear that changing the implementation class or spec.config is not an atomic operation and that any implementation would be likely to result in event loss during the transition.



# Testing Broker Lifecycle Conformance: 

We are going to be testing the previous two paragraphs coming from the Knative Eventing Spec. To do this we will be creating a broker checking its immutable properties, checking its Ready Status and then creating a Trigger that links to it by making a reference. We will also checking the Trigger Status, as it depends on the Broker to be ready to work correclty. We will be also checking that the broker is addresable by looking at the status conditions fields. Because this is a Control Plane test, we are not going to be sending Events to these components. 

You can find the resources for running these tests inside the `control-plane/broker-lifecycle/` directory. 
- A broker resource: `control-plane/broker-lifecycle/broker.yaml`
- A trigger resource that reference the broker: `control-plane/broker-lifecycle/trigger.yaml` 
- A trigger resource that doesn't reference the broker: `control-plane/broker-lifecycle/trigger-no-broker.yaml`


## [Pre] Creating a Broker 

```
kubectl apply -f control-plane/broker-lifecycle/broker.yaml
```


## [Test] Immutability

Check for default annotations, this should return the name of the selected implementation: 

```
kubectl get broker conformance-broker -ojson | jq '.metadata.annotations["eventing.knative.dev/broker.class"]'
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

**ISSUE Reported**: https://github.com/knative/eventing/issues/5663 

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
 kubectl get broker conformance-broker -ojson | jq '.status.conditions[] |select(.type == "Ready")' 
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
kubectl get broker conformance-broker -ojson | jq .status.address.url
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

## [Pre] Create Trigger for Broker

Create a trigger that points to the broker:

```
kubectl apply -f control-plane/broker-lifecycle/trigger.yaml
```

## [Test] Broker Reference in Trigger

Check that the `Trigger` is making a reference to the `Broker`, this should return the name of the broker.

```
kubectl get trigger conformance-trigger -ojson | jq '.spec.broker'
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

## [Test] Trigger for Broker Readyness

Check for condition type `Ready` with status `True`: 

```
kubectl get trigger conformance-trigger -ojson | jq '.status.conditions[] |select(.type == "Ready")'
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


# Clean up & Congrats

Make sure that you clean up all resources created in these tests by running: 

```
kubectl delete -f control-plane/broker-lifecycle/
```


Congratulations you have tested the **Broker Lifecycle Conformance** :metal: !
