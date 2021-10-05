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
   [data plane contract](./data-plane.md).
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

This set of tests provide conformance on the event delivery capabilities of Subscription/Channels and Trigger/Brokers resources. All referenced commands should be executed from the same directory this documentation is to be found.

It is important to clean all resources between each test listed here to avoid unexpected results.

## [Pre] Creating Channel, Broker, Role and RoleBinding

We will be testing a Channel and Broker using different Subscription and Trigger configurations. Test image [recordevents](https://github.com/knative/eventing/tree/main/test/test_images/recordevents) will be used for sending, receiving and logging events that assert conformance for each test. A Role and RoleBinding needed for the image is created in preparation for tests.

```
kubectl apply -f control-plane/event-delivery/00-prepare.yaml
```

## [Test] Channel delivery successful

- Create a Sink that successfuly receives and logs received events.
- Create a Subscription to the Channel.

```
kubectl apply -f control-plane/event-delivery/01-00-channel-ack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-ack-conformance ready: True
Pod/sink-ack-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/01-01-channel-ack.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~5 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"level":"warn","ts":1633423503.3684652,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633423503.3685536,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633423503.368569,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633424673.4930906,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-ack-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Length: 44\n  Content-Type: application/json\n  Traceparent: 00-681a141b86b0ff7070e4a627e3d70a25-d7a895a69311dc22-00\n  Host: sink-ack-conformance.conformance.svc.cluster.local\n  Prefer: reply\n  Accept-Encoding: gzip\n\n--- Origin: '172.17.0.1:18099' ---\n--- Observer: 'sink-ack-conformance' ---\n--- Time: 2021-10-05 09:04:33.492967905 +0000 UTC m=+1170.135547087 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/01-01-channel-ack.yaml,control-plane/event-delivery/01-00-channel-ack.yaml
```

## [Test] Broker delivery successful

- Create a Sink that successfuly receives and logs received events.
- Create a Trigger to the Broker.

```
kubectl apply -f control-plane/event-delivery/02-00-broker-ack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/sink-ack-conformance ready: True
Pod/sink-ack-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/02-01-broker-ack.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~5 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"level":"warn","ts":1633426277.6567748,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633426277.656887,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633426277.6569045,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633426423.961839,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T09:33:43.790423124Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-ack-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Length: 44\n  Traceparent: 00-e4eb717c66e0994b3ebd8a89ed9ada76-f027a004868f6ca9-00\n  Host: sink-ack-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:20744' ---\n--- Observer: 'sink-ack-conformance' ---\n--- Time: 2021-10-05 09:33:43.961638702 +0000 UTC m=+146.317709629 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/02-01-broker-ack.yaml,control-plane/event-delivery/02-00-broker-ack.yaml
```

## [Test] Channel delivery failed. No retries

- Create a Sink that does not ACK reception and logs received events.
- Create a Subscription to the Channel.

```
kubectl apply -f control-plane/event-delivery/03-00-channel-nack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-ack-conformance ready: True
Pod/sink-ack-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/03-01-channel-nack.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~5 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"level":"warn","ts":1633427213.464947,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633427213.4650588,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-nack-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:5}"}
{"level":"info","ts":1633427213.4650779,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633427369.1665733,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-nack-conformance\"\n  }\n\n--- HTTP headers ---\n  Prefer: reply\n  Host: sink-nack-conformance.conformance.svc.cluster.local\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Type: application/json\n  Traceparent: 00-bedf41bab7326588e9f799f407bd525f-e8e4963ac8671257-00\n  Content-Length: 45\n\n--- Origin: '172.17.0.1:55602' ---\n--- Observer: 'sink-nack-conformance' ---\n--- Time: 2021-10-05 09:49:29.166233581 +0000 UTC m=+155.713869892 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/03-01-channel-nack.yaml,control-plane/event-delivery/03-00-channel-nack.yaml
```

## [Test] Broker delivery failed. No retries

- Create a Sink that does not ACK reception and logs received events.
- Create a Trigger to the Broker.

```
kubectl apply -f control-plane/event-delivery/04-00-broker-nack.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/sink-nack-conformance ready: True
Pod/sink-nack-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/04-01-broker-nack.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~5 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"level":"warn","ts":1633427878.9139724,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633427878.9140756,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-nack-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:5}"}
{"level":"info","ts":1633427878.914092,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633428252.7688503,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T10:04:12.752758316Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-nack-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  Content-Length: 45\n  Host: sink-nack-conformance.conformance.svc.cluster.local\n  Traceparent: 00-86f47311bb7301cffab0e232341b4abb-278b2f69e8b103af-00\n  User-Agent: Go-http-client/1.1\n\n--- Origin: '172.17.0.1:53501' ---\n--- Observer: 'sink-nack-conformance' ---\n--- Time: 2021-10-05 10:04:12.768316902 +0000 UTC m=+373.866506110 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/04-00-broker-nack.yaml,control-plane/event-delivery/04-01-broker-nack.yaml
```

## [Test] Channel delivery failed + linear retries + successful.

- Create a Sink that fails to receive first 3 events, then succeeds, logging all received events.
- Create a Subscription to the Channel with delivery options set to 3 retries, 2 seconds backoff linear policy.

```
kubectl apply -f control-plane/event-delivery/05-00-channel-retry-linear.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-retry-linear-conformance ready: True
Pod/sink-retry-linear-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/05-01-channel-retry-linear.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~10 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly 4 entries for the received events that contain this EventInfo:

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
{"level":"warn","ts":1633429438.6321793,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633429438.6323245,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-retry-linear-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633429438.6323483,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633429568.9651825,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Length: 53\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n  Accept-Encoding: gzip\n  Content-Type: application/json\n  Prefer: reply\n  Traceparent: 00-5709f74332c8b6889913b9489dee7a10-3053ed2f857bb413-00\n  User-Agent: Go-http-client/1.1\n\n--- Origin: '172.17.0.1:10616' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:26:08.965096471 +0000 UTC m=+130.344498794 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633429568.9655774,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Length: 53\n  Accept-Encoding: gzip\n  Prefer: reply\n  Traceparent: 00-5709f74332c8b6889913b9489dee7a10-f92098cefcbd488b-00\n  User-Agent: Go-http-client/1.1\n  Content-Type: application/json\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:10616' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:26:08.965556392 +0000 UTC m=+130.344958707 ---\n--- Sequence: 2 ---\n--------------------\n"}
{"level":"info","ts":1633429570.9673288,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Traceparent: 00-5709f74332c8b6889913b9489dee7a10-c2ee426d7400dd02-00\n  User-Agent: Go-http-client/1.1\n  Content-Length: 53\n  Accept-Encoding: gzip\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n  Prefer: reply\n  Content-Type: application/json\n\n--- Origin: '172.17.0.1:10616' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:26:10.967240311 +0000 UTC m=+132.346642677 ---\n--- Sequence: 3 ---\n--------------------\n"}
{"level":"info","ts":1633429574.9680047,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Traceparent: 00-5709f74332c8b6889913b9489dee7a10-8bbced0bec42717a-00\n  Content-Type: application/json\n  Prefer: reply\n  Content-Length: 53\n  Accept-Encoding: gzip\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n  User-Agent: Go-http-client/1.1\n\n--- Origin: '172.17.0.1:10616' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:26:14.967978067 +0000 UTC m=+136.347380386 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/05-01-channel-retry-linear.yaml,control-plane/event-delivery/05-00-channel-retry-linear.yaml
```

## [Test] Broker delivery failed + linear retries + successful.

- Create a Sink that fails to receive first 3 events, then succeeds, logging all received events.
- Create a Trigger to the Broker with delivery options set to 3 retries, 2 seconds backoff linear policy.

```
kubectl apply -f control-plane/event-delivery/06-00-broker-retry-linear.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/sink-retry-linear-conformance ready: True
Pod/sink-retry-linear-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/06-01-broker-retry-linear.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~10 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly 4 entries for the received events that contain this EventInfo:

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
{"level":"warn","ts":1633430871.2206354,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633430871.2207692,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-retry-linear-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633430871.2207904,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633431033.4153495,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T10:50:33.399991955Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  User-Agent: Go-http-client/1.1\n  Content-Length: 53\n  Traceparent: 00-54820fe903d9c5b4ed58b047bbbc7a63-5eeebecd4ad49ab4-00\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n  Accept-Encoding: gzip\n\n--- Origin: '172.17.0.1:42309' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:50:33.415176667 +0000 UTC m=+162.210741081 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633431033.4170177,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T10:50:33.399991955Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  Traceparent: 00-54820fe903d9c5b4ed58b047bbbc7a63-95514e32adf631ba-00\n  Content-Length: 53\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:42309' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:50:33.416957989 +0000 UTC m=+162.212522404 ---\n--- Sequence: 2 ---\n--------------------\n"}
{"level":"info","ts":1633431035.4187968,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T10:50:33.399991955Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Traceparent: 00-54820fe903d9c5b4ed58b047bbbc7a63-ccb4dd960f19c9bf-00\n  Content-Type: application/json\n  User-Agent: Go-http-client/1.1\n  Content-Length: 53\n  Accept-Encoding: gzip\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:42309' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:50:35.418769049 +0000 UTC m=+164.214333454 ---\n--- Sequence: 3 ---\n--------------------\n"}
{"level":"info","ts":1633431039.4201734,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T10:50:33.399991955Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-linear-conformance\"\n  }\n\n--- HTTP headers ---\n  Host: sink-retry-linear-conformance.conformance.svc.cluster.local\n  Accept-Encoding: gzip\n  Content-Type: application/json\n  Traceparent: 00-54820fe903d9c5b4ed58b047bbbc7a63-03186dfb713b60c5-00\n  User-Agent: Go-http-client/1.1\n  Content-Length: 53\n\n--- Origin: '172.17.0.1:42309' ---\n--- Observer: 'sink-retry-linear-conformance' ---\n--- Time: 2021-10-05 10:50:39.420145519 +0000 UTC m=+168.215709916 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/06-01-broker-retry-linear.yaml,control-plane/event-delivery/06-00-broker-retry-linear.yaml
```

## [Test] Channel delivery failed + exponential retries + successful.

- Create a Sink that fails to receive first 3 events, then succeeds, logging all received events.
- Create a Subscription to the Channel with delivery options set to 3 retries, 2 seconds backoff exponential policy.

```
kubectl apply -f control-plane/event-delivery/07-00-channel-retry-exponential.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-retry-exponential-conformance ready: True
Pod/sink-retry-exponential-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/07-01-channel-retry-exponential.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~20 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly 4 entries for the received events that contain this EventInfo:

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
{"level":"warn","ts":1633432742.6688905,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633432742.6689892,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-retry-exponential-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633432742.6690052,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633432805.071768,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Prefer: reply\n  Content-Length: 58\n  Traceparent: 00-a9c5273ce18e2ab91baa6e530ed14516-9cf8ee9f2099a7ae-00\n  User-Agent: Go-http-client/1.1\n  Accept-Encoding: gzip\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:43100' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:20:05.071572956 +0000 UTC m=+62.414433410 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633432807.0734963,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Length: 58\n  Prefer: reply\n  Content-Type: application/json\n  Traceparent: 00-a9c5273ce18e2ab91baa6e530ed14516-65c6993e98db3b26-00\n  Accept-Encoding: gzip\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:43100' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:20:07.073446913 +0000 UTC m=+64.416307350 ---\n--- Sequence: 2 ---\n--------------------\n"}
{"level":"info","ts":1633432811.0748472,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  Accept-Encoding: gzip\n  Traceparent: 00-a9c5273ce18e2ab91baa6e530ed14516-2e9444dd0f1ed09d-00\n  User-Agent: Go-http-client/1.1\n  Content-Length: 58\n  Prefer: reply\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n  Content-Type: application/json\n\n--- Origin: '172.17.0.1:43100' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:20:11.074755694 +0000 UTC m=+68.417616169 ---\n--- Sequence: 3 ---\n--------------------\n"}
{"level":"info","ts":1633432819.076166,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Prefer: reply\n  Traceparent: 00-a9c5273ce18e2ab91baa6e530ed14516-f761ef7b87606415-00\n  Accept-Encoding: gzip\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n  User-Agent: Go-http-client/1.1\n  Content-Length: 58\n\n--- Origin: '172.17.0.1:43100' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:20:19.076130544 +0000 UTC m=+76.418990948 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/07-01-channel-retry-exponential.yaml,control-plane/event-delivery/07-00-channel-retry-exponential.yaml
```

## [Test] Broker delivery failed + exponential retries + successful.

- Create a Sink that fails to receive first 3 events, then succeeds, logging all received events.
- Create a Trigger to the Broker with delivery options set to 3 retries, 2 seconds backoff exponential policy.

```
kubectl apply -f control-plane/event-delivery/08-00-broker-retry-exponential.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/sink-retry-exponential-conformance ready: True
Pod/sink-retry-exponential-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/08-01-broker-retry-exponential.yaml
```

- Obtain logs for Sink (Interrupt output after logs stop, ~20 seconds)

```
kubectl logs -n conformance -f -l component=conformance-sink
```

### [Output]

Output must contain exactly 4 entries for the received events that contain this EventInfo:

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
{"level":"warn","ts":1633433434.3750818,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633433434.37518,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-retry-exponential-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633433434.3751962,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633433533.9549556,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T11:32:13.9420213Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  Content-Length: 58\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n  Traceparent: 00-87b60d68c1654bf7cd3d4f48bd55196e-3a7bfc5fd45df7ca-00\n  User-Agent: Go-http-client/1.1\n\n--- Origin: '172.17.0.1:38164' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:32:13.95478414 +0000 UTC m=+99.590775517 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633433535.9577682,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T11:32:13.9420213Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  Accept-Encoding: gzip\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n  Traceparent: 00-87b60d68c1654bf7cd3d4f48bd55196e-71de8bc436808ed0-00\n  User-Agent: Go-http-client/1.1\n  Content-Length: 58\n  Content-Type: application/json\n\n--- Origin: '172.17.0.1:38164' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:32:15.957673039 +0000 UTC m=+101.593664444 ---\n--- Sequence: 2 ---\n--------------------\n"}
{"level":"info","ts":1633433539.9605494,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T11:32:13.9420213Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Length: 58\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n  Content-Type: application/json\n  Traceparent: 00-87b60d68c1654bf7cd3d4f48bd55196e-a8411b2999a225d6-00\n\n--- Origin: '172.17.0.1:38164' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:32:19.960471959 +0000 UTC m=+105.596463349 ---\n--- Sequence: 3 ---\n--------------------\n"}
{"level":"info","ts":1633433547.9629848,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T11:32:13.9420213Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-retry-exponential-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Length: 58\n  Content-Type: application/json\n  Traceparent: 00-87b60d68c1654bf7cd3d4f48bd55196e-dfa4aa8dfbc4bcdb-00\n  Accept-Encoding: gzip\n  Host: sink-retry-exponential-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:38164' ---\n--- Observer: 'sink-retry-exponential-conformance' ---\n--- Time: 2021-10-05 11:32:27.962946227 +0000 UTC m=+113.598937597 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/08-01-broker-retry-exponential.yaml,control-plane/event-delivery/08-00-broker-retry-exponential.yaml
```

## [Test] Channel delivery failed + dead letter sink configured.

- Create a Sink that does not ACK reception and logs received events.
- Create secondary Sink that successfuly receives and logs received events.
- Create a Subscription to the Channel with delivery options for dead letter sink pointing to the secondary Sink.

```
kubectl apply -f control-plane/event-delivery/09-00-channel-dls.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-dls-conformance ready: True
Pod/sink-ack-dls-conformance ready: True
Pod/sink-nack-dls-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/09-01-channel-dls.yaml
```

- Obtain logs for both Sinks (Interrupt output after logs stop, ~5 seconds)

```
# First Sink NACKs incoming events
kubectl logs -n conformance -f -l app=sink-nack-dls-conformance

# Second Sink configured as dead letter sink ACKs incoming events
kubectl logs -n conformance -f -l app=sink-ack-dls-conformance
```

### [Output]

Output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"level":"warn","ts":1633434619.496375,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633434619.4965007,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-nack-dls-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633434619.4965205,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633434708.4816606,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-dls-conformance\"\n  }\n\n--- HTTP headers ---\n  Prefer: reply\n  Host: sink-nack-dls-conformance.conformance.svc.cluster.local\n  Content-Type: application/json\n  Traceparent: 00-9ebcd9e5036f07b19c1741e0e54ac449-089ef00fbcb69a49-00\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Length: 44\n\n--- Origin: '172.17.0.1:23467' ---\n--- Observer: 'sink-nack-dls-conformance' ---\n--- Time: 2021-10-05 11:51:48.481567021 +0000 UTC m=+88.999398793 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

Output for the second Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"level":"warn","ts":1633434607.0587378,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633434607.0588396,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-dls-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633434607.0588584,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633434708.484898,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativeerrorcode: 409\n  knativeerrordata: \n  knativeerrordest: http://sink-nack-dls-conformance.conformance.svc.cluster.local\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-dls-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Length: 44\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  Host: sink-ack-dls-conformance.conformance.svc.cluster.local\n  Traceparent: 00-9ebcd9e5036f07b19c1741e0e54ac449-9a39464dab3bc338-00\n\n--- Origin: '172.17.0.1:33957' ---\n--- Observer: 'sink-ack-dls-conformance' ---\n--- Time: 2021-10-05 11:51:48.484734962 +0000 UTC m=+101.439624297 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/09-01-channel-dls.yaml,control-plane/event-delivery/09-00-channel-dls.yaml
```

## [Test] Broker delivery failed + dead letter sink configured.

- Create a Sink that does not ACK reception and logs received events.
- Create secondary Sink that successfuly receives and logs received events.
- Create a Subscription to the Channel with delivery options for dead letter sink pointing to the secondary Sink.

```
kubectl apply -f control-plane/event-delivery/10-00-broker-dls.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/10-01-broker-dls.yaml
```

- Obtain logs for both Sinks (Interrupt output after logs stop, ~5 seconds)

```
# First Sink NACKs incoming events
kubectl logs -n conformance -f -l app=sink-nack-dls-conformance

# Second Sink configured as dead letter sink ACKs incoming events
kubectl logs -n conformance -f -l app=sink-ack-dls-conformance
```

### [Output]

Output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Rejected
- Sequence: 1

```
{"level":"warn","ts":1633435573.737146,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633435573.7373075,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-nack-dls-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633435573.737336,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633435658.2129655,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T12:07:38.198542474Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-dls-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Traceparent: 00-1921878d9b0e8cd7dacb237a7a9027d4-16083af25de753e1-00\n  Accept-Encoding: gzip\n  Content-Type: application/json\n  Content-Length: 44\n  Host: sink-nack-dls-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:25866' ---\n--- Observer: 'sink-nack-dls-conformance' ---\n--- Time: 2021-10-05 12:07:38.212775363 +0000 UTC m=+84.493123602 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

Output for the second Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"level":"warn","ts":1633435572.2848215,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633435572.2849698,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-dls-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633435572.284998,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633435658.2190907,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T12:07:38.198542474Z\n  knativebrokerttl: 255\n  knativeerrorcode: 409\n  knativeerrordata: \n  knativeerrordest: http://broker-filter.knative-eventing.svc.cluster.local/triggers/conformance/sink-retry-exponential-conformance/fb3e45fd-1519-4e73-a7ee-6974b0edf966\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-dls-conformance\"\n  }\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  Content-Length: 44\n  User-Agent: Go-http-client/1.1\n  Traceparent: 00-1921878d9b0e8cd7dacb237a7a9027d4-873e9c660188a88e-00\n  Host: sink-ack-dls-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:34144' ---\n--- Observer: 'sink-ack-dls-conformance' ---\n--- Time: 2021-10-05 12:07:38.218871598 +0000 UTC m=+85.951727653 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/10-01-broker-dls.yaml,control-plane/event-delivery/10-00-broker-dls.yaml
```

## [Test] Channel delivery + reply.

- Create a Sink that receives, replies and logs events.
- Create secondary Sink that receives and logs events.
- Create a Subscription to the Channel subscribed to first Sink and replying to second Sink.

```
kubectl apply -f control-plane/event-delivery/11-00-channel-reply.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-dls-conformance ready: True
Pod/sink-ack-conformance ready: True
Pod/sink-ack-reply-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/11-01-channel-reply.yaml
```

- Obtain logs for both Sinks (Interrupt output after logs stop, ~5 seconds)

```
# First Sink ACKs incoming events and produces a reply
kubectl logs -n conformance -f -l app=sink-ack-conformance

# Second Sink ACKs replies.
kubectl logs -n conformance -f -l app=sink-ack-reply-conformance
```

### [Output]

Output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

Information about composing the reply should be found after EventInfo entry.

```
{"level":"warn","ts":1633437717.2970679,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633437717.297206,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-conformance Reply:true ReplyEventType:conformance.reply ReplyEventSource:sink-ack-conformance ReplyEventData:conformance response message ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633437717.2972224,"logger":"fallback","caller":"receiver/receiver.go:90","msg":"Receiver will reply with an event"}
{"level":"info","ts":1633437754.8909645,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-reply-conformance\"\n  }\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Prefer: reply\n  Accept-Encoding: gzip\n  Content-Length: 46\n  Content-Type: application/json\n  Traceparent: 00-ce242df25c5784ab91d42ff915b70bf9-e2a79c42684f65f5-00\n  Host: sink-ack-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:13480' ---\n--- Observer: 'sink-ack-conformance' ---\n--- Time: 2021-10-05 12:42:34.890756639 +0000 UTC m=+37.606902723 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633437754.8910725,"logger":"fallback","caller":"receiver/reply.go:55","msg":"Setting reply event source 'sink-ack-conformance'"}
{"level":"info","ts":1633437754.8911397,"logger":"fallback","caller":"receiver/reply.go:59","msg":"Setting reply event type 'conformance.reply'"}
{"level":"info","ts":1633437754.891177,"logger":"fallback","caller":"receiver/reply.go:63","msg":"Setting reply event data ''"}
{"level":"info","ts":1633437754.8912323,"logger":"fallback","caller":"receiver/reply.go:84","msg":"Replying with","event":"Context Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData (binary),\n  conformance response message\n"}
```

Output for the second Sink must contain exactly one entry for the received reply event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"level":"warn","ts":1633437715.851353,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633437715.8514912,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-reply-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633437715.851515,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633437754.8995159,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  conformance response message\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Type: application/json\n  Traceparent: 00-ce242df25c5784ab91d42ff915b70bf9-7443f27f57d48de4-00\n  Accept-Encoding: gzip\n  Host: sink-ack-reply-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:13756' ---\n--- Observer: 'sink-ack-reply-conformance' ---\n--- Time: 2021-10-05 12:42:34.899375875 +0000 UTC m=+39.064342727 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Clean up]

```
kubectl delete -f control-plane/event-delivery/11-01-channel-reply.yaml,control-plane/event-delivery/11-00-channel-reply.yaml
```

## [Test] Broker delivery + reply.

- Create a Sink that receives, replies and logs events.
  - Replies should be typed `com.example.conformance`
- Create secondary Sink that receives and logs events.
- Create a Trigger to the Broker subscribed to first Sink
  - Trigger will filter for types `com.example.conformance`
- Create a Trigger and replying to second Sink.
  - Trigger will filter for types `com.example.conformance.reply`

```
kubectl apply -f control-plane/event-delivery/12-00-broker-reply.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance triggers,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Trigger/sink-ack-conformance ready: True
Trigger/sink-reply-conformance ready: True
Pod/sink-ack-conformance ready: True
Pod/sink-ack-reply-conformance ready: True
```

- Send single event typed `com.example.conformance`

```
kubectl apply -f control-plane/event-delivery/12-01-broker-reply.yaml
```

- Obtain logs for both Sinks (Interrupt output after logs stop, ~5 seconds)

```
# First Sink ACKs incoming events and produces a reply
kubectl logs -n conformance -f -l app=sink-ack-conformance

# Second Sink ACKs replies.
kubectl logs -n conformance -f -l app=sink-ack-reply-conformance
```

### [Output]

Output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

Information about composing the reply should be found after EventInfo entry.

```
{"level":"warn","ts":1633445075.1527643,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633445075.152873,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-conformance Reply:true ReplyEventType:com.example.conformance.reply ReplyEventSource:sink-ack-conformance ReplyEventData:conformance response message ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633445075.1528902,"logger":"fallback","caller":"receiver/receiver.go:90","msg":"Receiver will reply with an event"}
{"level":"info","ts":1633445249.05907,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T14:47:29.052680172Z\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-reply-conformance\"\n  }\n\n--- HTTP headers ---\n  Host: sink-ack-conformance.conformance.svc.cluster.local\n  Traceparent: 00-a4cde702478d640caa991c72c9b20a9b-244652413cf0787d-00\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Length: 46\n  Content-Type: application/json\n\n--- Origin: '172.17.0.1:45324' ---\n--- Observer: 'sink-ack-conformance' ---\n--- Time: 2021-10-05 14:47:29.058965627 +0000 UTC m=+173.920102087 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633445249.0591276,"logger":"fallback","caller":"receiver/reply.go:55","msg":"Setting reply event source 'sink-ack-conformance'"}
{"level":"info","ts":1633445249.0591486,"logger":"fallback","caller":"receiver/reply.go:59","msg":"Setting reply event type 'com.example.conformance.reply'"}
{"level":"info","ts":1633445249.059159,"logger":"fallback","caller":"receiver/reply.go:63","msg":"Setting reply event data ''"}
{"level":"info","ts":1633445249.0591671,"logger":"fallback","caller":"receiver/reply.go:84","msg":"Replying with","event":"Context Attributes,\n  specversion: 1.0\n  type: com.example.conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T14:47:29.052680172Z\n  sequence: 1\nData (binary),\n  conformance response message\n"}
```

Output for the second Sink must contain exactly one entry for the received reply event that contains this EventInfo:

- Kind: Received
- Sequence: 1

```
{"level":"warn","ts":1633445073.7351873,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633445073.7352839,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-reply-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633445073.7353072,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633445249.0654044,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  knativearrivaltime: 2021-10-05T14:47:29.062099454Z\n  sequence: 1\nData,\n  conformance response message\n\n--- HTTP headers ---\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Length: 28\n  Content-Type: application/json\n  Traceparent: 00-a4cde702478d640caa991c72c9b20a9b-d540413ee07e1fdc-00\n  Host: sink-ack-reply-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:42551' ---\n--- Observer: 'sink-ack-reply-conformance' ---\n--- Time: 2021-10-05 14:47:29.065270563 +0000 UTC m=+175.341490317 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Clean up]

```
kubectl delete -f control-plane/event-delivery/12-01-broker-reply.yaml,control-plane/event-delivery/12-00-broker-reply.yaml
```

## [Test] Channel delivery + reply with retries.

- Create a Sink that receives, replies and logs events.
- Create secondary Sink that fails to receive first 3 events, then succeeds, logging all received events.
- Create a Subscription to the Channel subscribed to first Sink and replying to second Sink, configuring delivery for 3 linear retries.

```
kubectl apply -f control-plane/event-delivery/13-00-channel-reply-retry.yaml
```

- Make sure all test objects are ready.

```
kubectl get -n conformance subscriptions,pods -ojsonpath="{range .items[*]}{@.kind}{'/'}{@.metadata.name}{' ready: '}{@.status.conditions[?(@.type=='Ready')].status}{'\n'}{end}"
```

- Expected output before proceeding with test.

```
Subscription/sink-reply-retry-conformance ready: True
Pod/sink-ack-conformance ready: True
Pod/sink-reply-retry-conformance ready: True
```

- Send single event.

```
kubectl apply -f control-plane/event-delivery/13-01-channel-reply-retry.yaml
```

- Obtain logs for both Sinks (Interrupt output after logs stop, ~5 seconds)

```
# First Sink ACKs incoming events and produces a reply
kubectl logs -n conformance -f -l app=sink-ack-conformance

# Second Sink ACKs replies.
kubectl logs -n conformance -f -l app=sink-reply-retry-conformance
```


### [Output]

Output for the first Sink must contain exactly one entry for the received event that contains this EventInfo:

- Kind: Received
- Sequence: 1

Information about composing the reply should be found after EventInfo entry.

```
{"level":"warn","ts":1633439183.0141628,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633439183.0142636,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-ack-conformance Reply:true ReplyEventType:conformance.reply ReplyEventSource:sink-ack-conformance ReplyEventData:conformance response message ReplyAppendData: SkipStrategy: SkipCounter:0}"}
{"level":"info","ts":1633439183.014283,"logger":"fallback","caller":"receiver/receiver.go:90","msg":"Receiver will reply with an event"}
{"level":"info","ts":1633439245.5785992,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: com.example.conformance\n  source: emitter-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  {\n    \"test\": \"sink-reply-retry-conformance\"\n  }\n\n--- HTTP headers ---\n  Prefer: reply\n  Traceparent: 00-6059795abd95c13ef1e4e470b01ee427-cfacf25bbe9b4a4b-00\n  Accept-Encoding: gzip\n  Content-Length: 52\n  User-Agent: Go-http-client/1.1\n  Host: sink-ack-conformance.conformance.svc.cluster.local\n  Content-Type: application/json\n\n--- Origin: '172.17.0.1:60845' ---\n--- Observer: 'sink-ack-conformance' ---\n--- Time: 2021-10-05 13:07:25.578330391 +0000 UTC m=+62.576214993 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633439245.5787084,"logger":"fallback","caller":"receiver/reply.go:55","msg":"Setting reply event source 'sink-ack-conformance'"}
{"level":"info","ts":1633439245.5787697,"logger":"fallback","caller":"receiver/reply.go:59","msg":"Setting reply event type 'conformance.reply'"}
{"level":"info","ts":1633439245.5788178,"logger":"fallback","caller":"receiver/reply.go:63","msg":"Setting reply event data ''"}
{"level":"info","ts":1633439245.5788631,"logger":"fallback","caller":"receiver/reply.go:84","msg":"Replying with","event":"Context Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData (binary),\n  conformance response message\n"}
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
{"level":"warn","ts":1633439181.5201247,"logger":"fallback","caller":"test_images/utils.go:80","msg":"Error while trying to read the config logging env: json logging string is empty"}
{"level":"info","ts":1633439181.5202482,"logger":"fallback","caller":"receiver/receiver.go:86","msg":"Receiver environment configuration: {ReceiverName:sink-reply-retry-conformance Reply:false ReplyEventType: ReplyEventSource: ReplyEventData: ReplyAppendData: SkipStrategy:sequence SkipCounter:3}"}
{"level":"info","ts":1633439181.5202672,"logger":"fallback","caller":"receiver/receiver.go:93","msg":"Receiver won't reply with an event"}
{"level":"info","ts":1633439245.5847304,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  conformance response message\n\n--- HTTP headers ---\n  Host: sink-reply-retry-conformance.conformance.svc.cluster.local\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Traceparent: 00-6059795abd95c13ef1e4e470b01ee427-61484899ad20733a-00\n\n--- Origin: '172.17.0.1:14069' ---\n--- Observer: 'sink-reply-retry-conformance' ---\n--- Time: 2021-10-05 13:07:25.584494404 +0000 UTC m=+64.077238454 ---\n--- Sequence: 1 ---\n--------------------\n"}
{"level":"info","ts":1633439245.5857763,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  conformance response message\n\n--- HTTP headers ---\n  User-Agent: Go-http-client/1.1\n  Content-Type: application/json\n  Traceparent: 00-6059795abd95c13ef1e4e470b01ee427-2a16f337256307b2-00\n  Host: sink-reply-retry-conformance.conformance.svc.cluster.local\n  Accept-Encoding: gzip\n\n--- Origin: '172.17.0.1:14069' ---\n--- Observer: 'sink-reply-retry-conformance' ---\n--- Time: 2021-10-05 13:07:25.585729355 +0000 UTC m=+64.078473416 ---\n--- Sequence: 2 ---\n--------------------\n"}
{"level":"info","ts":1633439247.5871377,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Rejected ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  conformance response message\n\n--- HTTP headers ---\n  Content-Type: application/json\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Traceparent: 00-6059795abd95c13ef1e4e470b01ee427-f3e39dd69ca59b29-00\n  Host: sink-reply-retry-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:14069' ---\n--- Observer: 'sink-reply-retry-conformance' ---\n--- Time: 2021-10-05 13:07:27.587106018 +0000 UTC m=+66.079850014 ---\n--- Sequence: 3 ---\n--------------------\n"}
{"level":"info","ts":1633439251.5910087,"logger":"fallback.event logger","caller":"logger_vent/logger.go:24","msg":"Event: \n-- EventInfo --\n--- Kind: Received ---\n--- Event ---\nContext Attributes,\n  specversion: 1.0\n  type: conformance.reply\n  source: sink-ack-conformance\n  id: 1\n  time: 2022-04-05T17:31:00Z\n  datacontenttype: application/json\nExtensions,\n  sequence: 1\nData,\n  conformance response message\n\n--- HTTP headers ---\n  Traceparent: 00-6059795abd95c13ef1e4e470b01ee427-bcb1487514e82fa1-00\n  Accept-Encoding: gzip\n  User-Agent: Go-http-client/1.1\n  Content-Type: application/json\n  Host: sink-reply-retry-conformance.conformance.svc.cluster.local\n\n--- Origin: '172.17.0.1:14069' ---\n--- Observer: 'sink-reply-retry-conformance' ---\n--- Time: 2021-10-05 13:07:31.590918822 +0000 UTC m=+70.083662865 ---\n--- Sequence: 1 ---\n--------------------\n"}
```

### [Cleanup]

```
kubectl delete -f control-plane/event-delivery/13-00-channel-reply-retry.yaml,control-plane/event-delivery/13-01-channel-reply-retry.yaml
```


# Clean up

Delete common resources for all event delivery tests.

```
kubectl delete -f control-plane/event-delivery/00-prepare.yaml
```
