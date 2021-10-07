# Event Delivery

From: https://github.com/knative/specs/blob/main/specs/eventing/control-plane.md#event-delivery

From the Spec:

>> Once a Trigger or Subscription has decided to deliver an event, it MUST do the
following:
>>
>> 1. Read the resolved URLs and delivery options from the object's `status`
   fields.
>>
>> 1. Attempt delivery to the `status.subscriberUri` URL following the
   [data plane contract](../../specs/eventing/control-plane.md).
>>
>>    1. If the event delivery fails with a retryable error, it MUST be retried up
      to `retry` times (subject to congestion control), following the
      `backoffPolicy` and `backoffDelay` parameters if specified.
>>
>> 1. If the delivery attempt is successful (either the original request or a
   retry) and no event is returned, the event delivery is complete.
>>
>> 1. If the delivery attempt is successful (either the original request or a
   retry) and an event is returned in the reply, the reply event MUST be
   delivered to the `status.replyUri` destination (for Subscriptions) or added
   to the Broker for processing (for Triggers). If `status.replyUri` is not
   present in the Subscription, the reply event MUST be dropped.
>>
>>    i. For Subscriptions, if delivery of the reply event fails with a retryable
      error, the entire delivery of the event to MUST be retried up to `retry`
      times (subject to congestion control), following the `backoffPolicy` and
      `backoffDelay` parameters if specified.
>>
>> 5. If an event (either the initial event or a reply) cannot be delivered, the
   event MUST be delivered to the `deadLetterSink` in the delivery options. If
   no `deadLetterSink` is specified, the event is dropped.
>>
>>    The implementation MAY set additional attributes on the event or wrap the
   failed event in a "failed delivery" event; this behavior is not (currently)
   standardized.
>>
>>    If delivery of the dead-letter event fails with a retryable error, the
   delivery to the `deadLetterSink` SHOULD be retried up to `retry` times,
   following the `backoffPolicy` and `backoffDelay` parameters if specified.
   Alternatively, implementations MAY use an equivalent internal mechanism for
   delivery (for example, if the `ref` form of `deadLetterSink` points to a
   compatible implementation).

# Testing Event Delivery Conformance

This set of tests provides conformance on the event delivery capabilities of Subscription/Channels and Trigger/Brokers resources. All referenced commands should be executed from the same directory this documentation is to be found.

You can find the resources for running these test inside the [control-plane/event-delivery/](control-plane/event-delivery/) directory

- [Base resources that include the Channel and Broker to be tested](control-plane/event-delivery/00-prepare.yaml).
- A pair of YAML files per test:
  - `<test-number>-00-<test-name>.yaml` for creating the testing scenario.
  - `<test-number>-01-<test-name>.yaml` for creating the event emmiter to start the test.

Tests are independent from each other and can be run in any other, but not concurrently. It is important to clean resources after running each test to avoid unexpected results.

## [Pre] Creating Channel, Broker, Role and RoleBinding

We will be testing a Channel and Broker using different Subscription and Trigger configurations. The test image [recordevents](https://github.com/knative/eventing/tree/main/test/test_images/recordevents) will be used for sending, receiving and logging events that assert conformance for each test. A Role and RoleBinding needed for the image is created in preparation for tests.

```
kubectl create -f control-plane/event-delivery/00-prepare.yaml
```

## [Test] Channel Delivery Successful

- Create a Sink that successfuly receives events.
- Create a Subscription to the Channel using the Sink.

```
kubectl create -f control-plane/event-delivery/01-00-channel-ack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-ack ready: True
Pod/sink-ack-qp7rt ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/01-01-channel-ack.yaml
```

### [Output]

The output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-ack"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["35"],"Content-Type":["application/json"],"Host":["sink-ack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-adbdc38a2bcfa28a5339c617102006f3-0d2c5a591ef84220-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:34793","observer":"sink-ack-hs4kb","time":"2021-10-06T17:59:03.035123383Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-ack
```

## [Test] Broker Delivery Successful

- Create a Sink that successfuly receives events.
- Create a Trigger to the Broker using the Sink as subscriber.

```
kubectl create -f control-plane/event-delivery/02-00-broker-ack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/broker-ack ready: True
Pod/sink-ack-m6768 ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/02-01-broker-ack.yaml
```

### [Output]

The output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-ack"},"sequence":"1","knativearrivaltime":"2021-10-06T18:12:44.077063771Z"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["34"],"Content-Type":["application/json"],"Host":["sink-ack.conformance.svc.cluster.local"],"Traceparent":["00-031c9fa2ecefc9069fdbc0a44d8ea12c-690a8cdffdb954b0-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:20749","observer":"sink-ack-m6768","time":"2021-10-06T18:12:44.096642003Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=broker-ack
```

## [Test] Channel Delivery Failed. No Retries

- Create a Sink that fails to receive events.
- Create a Subscription to the Channel using the Sink.

```
kubectl create -f control-plane/event-delivery/03-00-channel-nack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-nack ready: True
Pod/sink-3nack-vvnfb ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/03-01-channel-nack.yaml
```

### [Output]

The output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-nack"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["36"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-49f4884ea4c4a3179c933185d9cc7e8e-439b3b61e82453ff-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:20763","observer":"sink-3nack-vvnfb","time":"2021-10-06T18:17:45.988281885Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-nack
```

## [Test] Broker Delivery Failed. No Retries

- Create a Sink that fails to receive events.
- Create a Trigger to the Broker using the Sink as subscriber.

```
kubectl create -f control-plane/event-delivery/04-00-broker-nack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/broker-nack ready: True
Pod/sink-3nack-hb8bn ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/04-01-broker-nack.yaml
```

### [Output]

The output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-nack"},"sequence":"1","knativearrivaltime":"2021-10-06T20:15:34.641791303Z"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["35"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-78c480b1f8960bf8a700e646dd8c6d00-540ca19740351ddc-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:52158","observer":"sink-3nack-hb8bn","time":"2021-10-06T20:15:34.657542849Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=broker-nack
```

## [Test] Channel Delivery Failed + Linear Retries + Successful.

- Create a Sink that fails to receive first 3 events, then succeeds.
- Create a Subscription to the Channel with delivery options set to 3 retries and 2 seconds backoff linear policy.

```
kubectl create -f control-plane/event-delivery/05-00-channel-retry-linear.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-retry-linear ready: True
Pod/sink-3nack-47fnn ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~10 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/05-01-channel-retry-linear.yaml
```

Automated conformance test in Knative Eventing:
- https://github.com/knative/eventing/blob/27dbb99773a0db7860d14f34df65321e145aaf48/test/rekt/features/channel/data_plane.go#L136-L146

### [Output]

The output must contain exactly 4 entries for the received events that contain this EventInfo:

- Event 1:
  - Kind: Rejected
  - Sequence: 1
  - Time: \<first event arrival time\>
- Event 2:
  - Kind: Rejected
  - Sequence: 2
  - Time: \<previous event time\> + roundtrip delay.
- Event 3:
  - Kind: Rejected
  - Sequence: 3
  - Time: \<previous event time\> + roundtrip delay + 2 seconds.
- Event 4:
  - Kind: Received
  - Sequence: 1
  - Time: \<previous event time\> + roundtrip delay + 4 seconds.

Roundtrip delay is expected to be a similar ammount of time for all received events, and ideally should be far from the 1s range.

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-linear"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["44"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-7a2d55d8de36d78f8ee6b45bc8732a60-790a1d69b25163de-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:7305","observer":"sink-3nack-47fnn","time":"2021-10-06T20:19:25.843310855Z","sequence":1}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-linear"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["44"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-7a2d55d8de36d78f8ee6b45bc8732a60-029dc2bf7eae9003-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:7305","observer":"sink-3nack-47fnn","time":"2021-10-06T20:19:25.866810367Z","sequence":2}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-linear"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["44"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-7a2d55d8de36d78f8ee6b45bc8732a60-8b2f68164b0bbe28-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:7305","observer":"sink-3nack-47fnn","time":"2021-10-06T20:19:27.874018224Z","sequence":3}
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-linear"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["44"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-7a2d55d8de36d78f8ee6b45bc8732a60-14c20d6d1768eb4d-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:7305","observer":"sink-3nack-47fnn","time":"2021-10-06T20:19:31.882981554Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-retry-linear
```

## [Test] Broker Delivery Failed + Linear Retries + Successful.

- Create a Sink that fails to receive first 3 events, then succeeds.
- Create a Trigger to the Broker with delivery options set to 3 retries and 2 seconds backoff linear policy.

```
kubectl create -f control-plane/event-delivery/06-00-broker-retry-linear.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/broker-retry-linear ready: True
Pod/sink-3nack-jcv2v ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~10 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/06-01-broker-retry-linear.yaml
```

### [Output]

The output must contain exactly 4 entries for the received events that contain this EventInfo:

- Event 1:
  - Kind: Rejected
  - Sequence: 1
  - Time: \<first event arrival time\>
- Event 2:
  - Kind: Rejected
  - Sequence: 2
  - Time: \<previous event time\> + roundtrip delay.
- Event 3:
  - Kind: Rejected
  - Sequence: 3
  - Time: \<previous event time\> + roundtrip delay + 2 seconds.
- Event 4:
  - Kind: Received
  - Sequence: 1
  - Time: \<previous event time\> + roundtrip delay + 4 seconds.

Roundtrip delay is expected to be a similar ammount of time for all received events, and ideally should be far from the 1s range.

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-linear"},"sequence":"1","knativearrivaltime":"2021-10-06T20:22:52.32099058Z"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["43"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-514aa7bb1a9cbe59fc01bfb3684c1f2d-3f0eb64f83b0e507-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:32643","observer":"sink-3nack-jcv2v","time":"2021-10-06T20:22:52.329323084Z","sequence":1}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-linear"},"sequence":"1","knativearrivaltime":"2021-10-06T20:22:52.32099058Z"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["43"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-514aa7bb1a9cbe59fc01bfb3684c1f2d-2a10cb07c62bae33-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:32643","observer":"sink-3nack-jcv2v","time":"2021-10-06T20:22:52.336212658Z","sequence":2}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-linear"},"knativearrivaltime":"2021-10-06T20:22:52.32099058Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["43"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-514aa7bb1a9cbe59fc01bfb3684c1f2d-1512e0bf08a7765f-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:32643","observer":"sink-3nack-jcv2v","time":"2021-10-06T20:22:54.341294268Z","sequence":3}
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-linear"},"knativearrivaltime":"2021-10-06T20:22:52.32099058Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["43"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-514aa7bb1a9cbe59fc01bfb3684c1f2d-0014f5774b223f8b-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:32643","observer":"sink-3nack-jcv2v","time":"2021-10-06T20:22:58.354717624Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=broker-retry-linear
```

## [Test] Channel Delivery Failed + Exponential Retries + Successful.

- Create a Sink that fails to receive first 3 events, then succeeds.
- Create a Subscription to the Channel with delivery options set to 3 retries and 2 seconds backoff exponential policy.

```
kubectl create -f control-plane/event-delivery/07-00-channel-retry-exponential.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-retry-exponential ready: True
Pod/sink-3nack-dcq4q ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~20 seconds for the event to be reported at the watched output.

```
kubectl create -f control-plane/event-delivery/07-01-channel-retry-exponential.yaml
```

### [Output]

The output must contain exactly 4 entries for the received events that contain this EventInfo:

- Event 1:
  - Kind: Rejected
  - Sequence: 1
  - Time: \<first event arrival time\>
- Event 2:
  - Kind: Rejected
  - Sequence: 2
  - Time: \<previous event time\> + roundtrip delay + 2 seconds.
- Event 3:
  - Kind: Rejected
  - Sequence: 3
  - Time: \<previous event time\> + roundtrip delay + 4 seconds.
- Event 4:
  - Kind: Received
  - Sequence: 1
  - Time: \<previous event time\> + roundtrip delay + 8 seconds.

Roundtrip delay is expected to be a similar ammount of time for all received events, and ideally should be far from the 1s range.

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-exponential"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["49"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-f42bbc14a1b4922085e559f2f8c2eb50-e5e8df7846ab839c-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:53630","observer":"sink-3nack-dcq4q","time":"2021-10-06T20:37:07.838555942Z","sequence":1}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-exponential"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["49"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-f42bbc14a1b4922085e559f2f8c2eb50-6e7b85cf1208b1c1-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:53630","observer":"sink-3nack-dcq4q","time":"2021-10-06T20:37:09.865454188Z","sequence":2}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-exponential"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["49"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-f42bbc14a1b4922085e559f2f8c2eb50-f70d2b26df64dee6-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:53630","observer":"sink-3nack-dcq4q","time":"2021-10-06T20:37:13.878609916Z","sequence":3}
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-retry-exponential"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["49"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-f42bbc14a1b4922085e559f2f8c2eb50-80a0d07cabc10b0c-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:53630","observer":"sink-3nack-dcq4q","time":"2021-10-06T20:37:21.890764102Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-retry-exponential
```

## [Test] Broker Delivery Failed + Exponential Retries + Successful.

- Create a Sink that fails to receive first 3 events, then succeeds.
- Create a Trigger to the Broker with delivery options set to 3 retries and 2 seconds backoff exponential policy.

```
kubectl create -f control-plane/event-delivery/08-00-broker-retry-exponential.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/broker-retry-exponential ready: True
Pod/sink-3nack-nqqx5 ready: True
```

- Create a new shell and watch for the events reported by the Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~20 seconds for the event to be reported at the watched output.


- Send single event.

```
kubectl create -f control-plane/event-delivery/08-01-broker-retry-exponential.yaml
```

### [Output]

The output must contain exactly 4 entries for the received events that contain this EventInfo:

- Event 1:
  - Kind: Rejected
  - Sequence: 1
  - Time: \<first event arrival time\>
- Event 2:
  - Kind: Rejected
  - Sequence: 2
  - Time: \<previous event time\> + roundtrip delay + 2 seconds.
- Event 3:
  - Kind: Rejected
  - Sequence: 3
  - Time: \<previous event time\> + roundtrip delay + 4 seconds.
- Event 4:
  - Kind: Received
  - Sequence: 1
  - Time: \<previous event time\> + roundtrip delay + 8 seconds.

Roundtrip delay is expected to be a similar ammount of time for all received events, and ideally should be far from the 1s range.

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-exponential"},"knativearrivaltime":"2021-10-06T20:40:47.231977908Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["48"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-fa43406b48ae1fccd92125e3e9d0657d-eb150a308e9d07b7-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:47537","observer":"sink-3nack-nqqx5","time":"2021-10-06T20:40:47.237818522Z","sequence":1}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-exponential"},"knativearrivaltime":"2021-10-06T20:40:47.231977908Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["48"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-fa43406b48ae1fccd92125e3e9d0657d-d6171fe8d018d0e2-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:47537","observer":"sink-3nack-nqqx5","time":"2021-10-06T20:40:49.246108503Z","sequence":2}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-exponential"},"knativearrivaltime":"2021-10-06T20:40:47.231977908Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["48"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-fa43406b48ae1fccd92125e3e9d0657d-c11934a01394980e-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:47537","observer":"sink-3nack-nqqx5","time":"2021-10-06T20:40:53.2573393Z","sequence":3}
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-retry-exponential"},"knativearrivaltime":"2021-10-06T20:40:47.231977908Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["48"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-fa43406b48ae1fccd92125e3e9d0657d-ac1b4958560f613a-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:47537","observer":"sink-3nack-nqqx5","time":"2021-10-06T20:41:01.265325854Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=broker-retry-exponential
```

## [Test] Channel Delivery Failed + Dead Letter Sink Configured.

- Create a Sink that fails to receive events.
- Create secondary Sink that successfuly receives events.
- Create a Subscription to the Channel using the first Sink with Dead Letter Sink pointing to the secondary Sink.

```
kubectl create -f control-plane/event-delivery/09-00-channel-dls.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-dls ready: True
Pod/sink-3nack-w9wbc ready: True
Pod/sink-ack-zfcw8 ready: True
```
- Create two new shell, at the first one watch for the events reported by the first rejecting Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- At the second shell watch for the events sent to the Dead Letter Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```


- Send single event, then wait ~5 seconds for the events to be reported at both outputs.

```
kubectl create -f control-plane/event-delivery/09-01-channel-dls.yaml
```

Automated conformance test in Knative Eventing:
- https://github.com/knative/eventing/blob/27dbb99773a0db7860d14f34df65321e145aaf48/test/rekt/features/channel/data_plane.go#L170-L179

### [Output]

The output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-dls"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["35"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-fce81158b4f47711c69266415b29ea23-51c7a288da04a45a-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:9275","observer":"sink-3nack-w9wbc","time":"2021-10-06T20:49:11.219432351Z","sequence":1}
```

Output for the second Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-dls"},"knativeerrorcode":"409","sequence":"1","knativeerrordata":"","knativeerrordest":"http://sink-3nack.conformance.svc.cluster.local"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["35"],"Content-Type":["application/json"],"Host":["sink-ack.conformance.svc.cluster.local"],"Traceparent":["00-fce81158b4f47711c69266415b29ea23-63eced3573befea4-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:31746","observer":"sink-ack-zfcw8","time":"2021-10-06T20:49:11.246469118Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-dls
```

## [Test] Broker Delivery Failed + Dead Letter Sink Configured.

- Create a Sink that fails to receive events.
- Create secondary Sink that successfuly receives events.
- Create a Trigger at the Broker using the first Sink with Dead Letter Sink pointing to the secondary Sink.

```
kubectl create -f control-plane/event-delivery/10-00-broker-dls.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/broker-dls ready: True
Pod/sink-3nack-vmwl4 ready: True
Pod/sink-ack-r8bv9 ready: True
```

- Create two new shell, at the first one watch for the events reported by the first rejecting Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- At the second shell watch for the events sent to the Dead Letter Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the events to be reported at both outputs.


```
kubectl create -f control-plane/event-delivery/10-01-broker-dls.yaml
```

### [Output]

The output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-dls"},"knativearrivaltime":"2021-10-06T20:55:19.451156218Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["34"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-9a7177177d05434f8f0c13a60eb9f8d1-971d5e10998a2966-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:5575","observer":"sink-3nack-vmwl4","time":"2021-10-06T20:55:19.465001286Z","sequence":1}
```

Output for the second Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-dls"},"sequence":"1","knativebrokerttl":"255","knativeerrordest":"http://broker-filter.knative-eventing.svc.cluster.local/triggers/conformance/broker-dls/c75b9428-4663-4328-b2bf-f91eb6bc4df0","knativeerrordata":"","knativearrivaltime":"2021-10-06T20:55:19.451156218Z","knativeerrorcode":"409"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["34"],"Content-Type":["application/json"],"Host":["sink-ack.conformance.svc.cluster.local"],"Traceparent":["00-9a7177177d05434f8f0c13a60eb9f8d1-10c929e7708ee15e-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:2618","observer":"sink-ack-r8bv9","time":"2021-10-06T20:55:19.49591768Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=broker-dls
```

## [Test] Channel Delivery + Reply.

- Create a Sink that receives events and replies with a new event.
- Create secondary Sink that receives events.
- Create a Subscription to the Channel using the first and replying to second Sink.

```
kubectl create -f control-plane/event-delivery/11-00-channel-reply.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-reply ready: True
Pod/sink-ack-jdfbc ready: True
Pod/sink-ack-reply-pf6fd ready: True
```

- Create two new shell, at the first one watch for the events reported by the first Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack-reply -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- At the second shell watch for the events sent as a response to the second Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the events to be reported at both outputs.

```
kubectl create -f control-plane/event-delivery/11-01-channel-reply.yaml
```

Automated conformance test in Knative Eventing:
- https://github.com/knative/eventing/blob/27dbb99773a0db7860d14f34df65321e145aaf48/test/rekt/features/channel/data_plane.go#L148-L157

### [Output]

The output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

Information about composing the reply should be found after EventInfo entry.

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-reply"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["37"],"Content-Type":["application/json"],"Host":["sink-ack-reply.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-e46e0d78021c184dced1d68916752bea-1c68b6750ebfe002-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:59083","observer":"sink-ack-reply-g8t6h","time":"2021-10-06T21:48:23.087506617Z","sequence":1}
```

Output for the second Sink must contain exactly one entry for the received reply event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"sink-ack-reply-g8t6h","type":"conformance.reply","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"conformance":"response message"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Type":["application/json"],"Host":["sink-ack.conformance.svc.cluster.local"],"Traceparent":["00-e46e0d78021c184dced1d68916752bea-5e5de0f603a9c16d-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:15854","observer":"sink-ack-4f7r7","time":"2021-10-06T21:48:23.101716012Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-reply
```

## [Test] Broker Delivery + Reply.

- Create a Sink that receives events and replies with a new event typed `com.example.conformance.reply`.
- Create secondary Sink that receives events.
- Create a Trigger to the Broker subscribed to the first Sink, set an attribute filter for type `com.example.conformance`
- Create a Trigger to the Broker subscribed to the second Sink, set an attribute filter for type `com.example.conformance.reply`

```
kubectl create -f control-plane/event-delivery/12-00-broker-reply.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/sink-ack ready: True
Trigger/sink-ack-reply ready: True
Pod/sink-ack-reply-j2cf6 ready: True
Pod/sink-ack-vm9jn ready: True
```

- Create two new shell, at the first one watch for the events reported by the first Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack-reply -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- At the second shell watch for the response events reported by the second sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~5 seconds for the events to be reported at both outputs.

- Send single event typed `com.example.conformance`

```
kubectl create -f control-plane/event-delivery/12-01-broker-reply.yaml
```

### [Output]

The output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

Information about composing the reply should be found after EventInfo entry.

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"broker-reply"},"knativearrivaltime":"2021-10-06T22:47:43.888215638Z","sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["36"],"Content-Type":["application/json"],"Host":["sink-ack-reply.conformance.svc.cluster.local"],"Traceparent":["00-f42f0ecf02ee71223da07e734fcf89d7-24ea7bc4330caa6b-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:52698","observer":"sink-ack-reply-njng5","time":"2021-10-06T22:47:43.894345725Z","sequence":1}
```

Output for the second Sink must contain exactly one entry for the received reply event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"sink-ack-reply-njng5","type":"com.example.conformance.reply","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"conformance":"response message"},"sequence":"1","knativearrivaltime":"2021-10-06T22:47:43.902529577Z"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["34"],"Content-Type":["application/json"],"Host":["sink-ack.conformance.svc.cluster.local"],"Traceparent":["00-f42f0ecf02ee71223da07e734fcf89d7-d79b736728d0330c-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:14904","observer":"sink-ack-kfkd6","time":"2021-10-06T22:47:43.904828124Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=broker-reply
```

## [Test] Channel Delivery + Reply with Retries.

- Create a Sink that receives events and replies with a new event.
- Create secondary Sink that fails to receive first 3 events, then succeeds.
- Create a Subscription to the Channel using the first and replying to the second Sink, configuring delivery for 3 linear retries.

```
kubectl create -f control-plane/event-delivery/13-00-channel-reply-retry.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/channel-reply-retry ready: True
Pod/sink-3nack-q6t2v ready: True
Pod/sink-ack-reply-nkb9n ready: True
```

- Open two shells. At the first one watch the output for events reported by the first Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-ack-reply -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- At the second shell watch the output for events sent as a response to the second Sink.

```
kubectl get events -n conformance --field-selector reason=CloudEventObserved,involvedObject.name=$(kubectl get pods -n conformance -l component=sink-3nack -o jsonpath='{.items[0].metadata.name}') -w --output=custom-columns=:.message
```

- Send single event, then wait ~10 seconds for the events to be reported at both outputs.

```
kubectl create -f control-plane/event-delivery/13-01-channel-reply-retry.yaml
```

Automated conformance test in Knative Eventing:
- https://github.com/knative/eventing/blob/27dbb99773a0db7860d14f34df65321e145aaf48/test/rekt/features/channel/data_plane.go#L159-L167

### [Output]

The output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

Information about composing the reply should be found after EventInfo entry.

```
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"emitter-conformance","type":"com.example.conformance","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"test":"channel-reply-retry"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Length":["43"],"Content-Type":["application/json"],"Host":["sink-ack-reply.conformance.svc.cluster.local"],"Prefer":["reply"],"Traceparent":["00-0b896abb42ee4d5105f76f9ae9d34e30-aeaf6e256037e620-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:21770","observer":"sink-ack-reply-nkb9n","time":"2021-10-06T22:53:52.616010111Z","sequence":1}
```

Output for the second Sink must contain exactly 4 entries for the received reply event that contains this EventInfo:

- Event 1:
  - Kind: Rejected
  - Sequence: 1
  - Time: \<first event arrival time\>
- Event 2:
  - Kind: Rejected
  - Sequence: 2
  - Time: \<previous event time\> + roundtrip delay.
- Event 3:
  - Kind: Rejected
  - Sequence: 3
  - Time: \<previous event time\> + roundtrip delay + 2 seconds.
- Event 4:
  - Kind: Received
  - Sequence: 1
  - Time: \<previous event time\> + roundtrip delay + 4 seconds.

Roundtrip delay is expected to be a similar ammount of time for all received events, and ideally should be far from the 1s range.

```
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"sink-ack-reply-nkb9n","type":"conformance.reply","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"conformance":"response message"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-0b896abb42ee4d5105f76f9ae9d34e30-f0a498a65521c78b-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:29073","observer":"sink-3nack-q6t2v","time":"2021-10-06T22:53:52.640394823Z","sequence":1}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"sink-ack-reply-nkb9n","type":"conformance.reply","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"conformance":"response message"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-0b896abb42ee4d5105f76f9ae9d34e30-919f2d6750963741-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:29073","observer":"sink-3nack-q6t2v","time":"2021-10-06T22:53:52.660831023Z","sequence":2}
{"kind":"Rejected","event":{"specversion":"1.0","id":"1","source":"sink-ack-reply-nkb9n","type":"conformance.reply","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"conformance":"response message"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-0b896abb42ee4d5105f76f9ae9d34e30-329ac2274b0ba8f6-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:29073","observer":"sink-3nack-q6t2v","time":"2021-10-06T22:53:54.667649576Z","sequence":3}
{"kind":"Received","event":{"specversion":"1.0","id":"1","source":"sink-ack-reply-nkb9n","type":"conformance.reply","datacontenttype":"application/json","time":"2022-04-05T17:31:00Z","data":{"conformance":"response message"},"sequence":"1"},"httpHeaders":{"Accept-Encoding":["gzip"],"Content-Type":["application/json"],"Host":["sink-3nack.conformance.svc.cluster.local"],"Traceparent":["00-0b896abb42ee4d5105f76f9ae9d34e30-d39457e8458018ac-00"],"User-Agent":["Go-http-client/1.1"]},"origin":"172.17.0.1:29073","observer":"sink-3nack-q6t2v","time":"2021-10-06T22:53:58.677521753Z","sequence":1}
```

### [Clean up]

```
kubectl delete all -n conformance -l case=channel-reply-retry
```

# Clean up

Delete common resources for all event delivery tests.

```
kubectl delete -f control-plane/event-delivery/00-prepare.yaml
```
