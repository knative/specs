# Error Signalling

#### From:

https://github.com/knative/specs/blob/main/specs/common/error-signalling.md

#### From the Spec:

> [...] Each user-visible resource described in [Resource
> Overview][res-overview] MUST have a `conditions` field in `status`, which MUST
> be a list of `Condition` objects described by the following table.
>
> Fields in the condition which are not marked as "REQUIRED" MAY be omitted to
> indicate the default value [...]. As `Conditions` are an output API, an
> implementation MAY never set these fields; the OpenAPI document MUST still
> describe these fields.

> Additionally, the resource's `status.conditions` field MUST be managed as
> follows to enable clients (particularly user interfaces) to present useful
> diagnostic and error message to the user. In the following section, conditions
> are referred to by their `type` (aka the string value of the `type` field on
> the Condition).
>
> 1.  Each resource MUST have a summary `Ready` condition (for ongoing systems)
>     or `Succeeded` condition (for resources that run to completion) with
>     `severity=""`, which MUST use the `"True"`, `"False"`, and `"Unknown"`
>     status values as follows:
>
>     1.  `"False"` MUST indicate a failure condition.
>
>     [...]
>
> 1.  For non-`Ready` conditions, any conditions with `severity=""` (aka "Error
>     conditions") MUST be aggregated into the "Ready" condition as follows:
>
>     1.  If the condition is `"False"`, `Ready`'s status MUST be `"False"`.
>     1.  If the condition is `"Unknown"`, `Ready`'s status MUST be `"False"` or
>         `"Unknown"`.
>     1.  If the condition is `"True"`, `Ready`'s status can be any of `"True"`,
>         `"False"`, or `"Unknown"`.
>
>     Implementations MAY choose to report that `Ready` is `"False"` or
>     `"Unknown"` even if all Error conditions report a status of `"True"` (i.e.
>     there might be additional hidden implementation conditions which feed into
>     the `Ready` condition which are not reported.)
>
> 1.  Conditions with a `status` other than `"True"` SHOULD provide `message`
>     and `reason` fields indicating the reason that the `status` is not
>     `"True"`. Conditions where the `status` is `"False"` MUST provide a
>     failure `reason` in the condition. (`"Unknown"` conditions might not have
>     been reconciled, and so MAY have an empty `reason`.)

[res-overview]: https://github.com/knative/specs/blob/c348f501/specs/eventing/overview.md

# Testing Error Signalling Conformance:

We are going to be testing that each of the core resources in the eventing and
messaging API groups - Broker, Trigger, Channel and Subscription - satisfies the
aforementioned parts of the Knative Eventing spec. We will achieve this by:

- Asserting that the OpenAPI schema declared inside the CustomResourceDefinition
  of each resource contains all required status fields.

- Creating various instances of each resource, causing them to transition
  through different states of success and failure, and asserting that their
  status is propagated accordingly.


The resources necessary for running these tests are:

- `test-plan/eventing/control-plane/error-signalling/broker-not-ready.yaml`
- `test-plan/eventing/control-plane/error-signalling/trigger-not-ready.yaml`
- `test-plan/eventing/control-plane/error-signalling/channel-not-ready.yaml`
- `test-plan/eventing/control-plane/error-signalling/subscription-not-ready.yaml`

---

## [Test] OpenAPI schema for the Broker kind

Describe the registered `brokers.eventing.knative.dev` kind and verify that it
contains a `status.conditions` field, expressed as a list of objects which
themselves exclusively contain the fields:

- `type`, typed as a string and marked as required
- `status`, typed as a string
- `message` typed as a string
- `severity`, typed as a string
- `lastTransitionTime`, typed as a string

Explain the `status.conditions` field:

```
kubectl explain brokers.eventing.knative.dev.status.conditions
```

### [Output]

```
KIND:     Broker
VERSION:  eventing.knative.dev/v1

RESOURCE: conditions <[]Object>

[...]

FIELDS:
   lastTransitionTime   <string>
     LastTransitionTime is the last time the condition transitioned from one
     status to another. We use VolatileTime in place of metav1.Time to exclude
     this from creating equality.Semantic differences (all other things held
     constant).

   message      <string>
     A human readable message indicating details about the transition.

   reason       <string>
     The reason for the condition's last transition.

   severity     <string>
     Severity with which to treat failures of this type of condition. When this
     is not specified, it defaults to Error.

   status       <string> -required-
     Status of the condition, one of True, False, Unknown.

   type <string> -required-
     Type of condition.
```

---

## [Test] OpenAPI schema for the Trigger kind

Describe the registered `triggers.eventing.knative.dev` kind and verify that it
contains a `status.conditions` field, expressed as a list of objects which
themselves exclusively contain the fields:

- `type`, typed as a string and marked as required
- `status`, typed as a string
- `message` typed as a string
- `severity`, typed as a string
- `lastTransitionTime`, typed as a string

Explain the `status.conditions` field:

```
kubectl explain triggers.eventing.knative.dev.status.conditions
```

### [Output]

```
KIND:     Trigger
VERSION:  eventing.knative.dev/v1

RESOURCE: conditions <[]Object>

[...]

FIELDS:
   lastTransitionTime   <string>
     LastTransitionTime is the last time the condition transitioned from one
     status to another. We use VolatileTime in place of metav1.Time to exclude
     this from creating equality.Semantic differences (all other things held
     constant).

   message      <string>
     A human readable message indicating details about the transition.

   reason       <string>
     The reason for the condition's last transition.

   severity     <string>
     Severity with which to treat failures of this type of condition. When this
     is not specified, it defaults to Error.

   status       <string> -required-
     Status of the condition, one of True, False, Unknown.

   type <string> -required-
     Type of condition.
```

---

## [Test] OpenAPI schema for the Channel kind

Describe the registered `channels.messaging.knative.dev` kind and verify that
it contains a `status.conditions` field, expressed as a list of objects which
themselves exclusively contain the fields:

- `type`, typed as a string and marked as required
- `status`, typed as a string
- `message` typed as a string
- `severity`, typed as a string
- `lastTransitionTime`, typed as a string

Explain the `status` field:

```
kubectl explain channels.messaging.knative.dev.status.conditions
```

### [Output]

```
KIND:     Channel
VERSION:  messaging.knative.dev/v1

RESOURCE: conditions <[]Object>

[...]

FIELDS:
   lastTransitionTime   <string>
     LastTransitionTime is the last time the condition transitioned from one
     status to another. We use VolatileTime in place of metav1.Time to exclude
     this from creating equality.Semantic differences (all other things held
     constant).

   message      <string>
     A human readable message indicating details about the transition.

   reason       <string>
     The reason for the condition's last transition.

   severity     <string>
     Severity with which to treat failures of this type of condition. When this
     is not specified, it defaults to Error.

   status       <string> -required-
     Status of the condition, one of True, False, Unknown.

   type <string> -required-
     Type of condition.
```

---

## [Test] OpenAPI schema for the Subscription kind

Describe the registered `subscriptions.messaging.knative.dev` kind and verify
that it contains a `status.conditions` field, expressed as a list of objects
which themselves exclusively contain the fields:

- `type`, typed as a string and marked as required
- `status`, typed as a string
- `message` typed as a string
- `severity`, typed as a string
- `lastTransitionTime`, typed as a string

Explain the `status` field:

```
kubectl explain subscriptions.messaging.knative.dev.status.conditions
```

### [Output]

```
KIND:     Subscription
VERSION:  messaging.knative.dev/v1

RESOURCE: conditions <[]Object>

[...]

FIELDS:
   lastTransitionTime   <string>
     LastTransitionTime is the last time the condition transitioned from one
     status to another. We use VolatileTime in place of metav1.Time to exclude
     this from creating equality.Semantic differences (all other things held
     constant).

   message      <string>
     A human readable message indicating details about the transition.

   reason       <string>
     The reason for the condition's last transition.

   severity     <string>
     Severity with which to treat failures of this type of condition. When this
     is not specified, it defaults to Error.

   status       <string> -required-
     Status of the condition, one of True, False, Unknown.

   type <string> -required-
     Type of condition.
```

---

## [Pre] Create a Broker instance which reconciliation fails

```
kubectl create -f control-plane/error-signalling/broker-not-ready.yaml
```

## [Test] Broker failure is propagated to the Ready status condition

The failure caused by the intentionally invalid spec is propagated to the `Ready` condition, which has a `status` of
`False`, a severity of `""` (error), and a non-empty `reason`:

```
kubectl get brokers/not-ready -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}{";"}{.status.conditions[?(@.type=="Ready")].severity}{";"}{.status.conditions[?(@.type=="Ready")].reason}'
```

### [Output]

```
False;;ChannelTemplateFailed
```

---

## [Pre] Create a Trigger instance which reconciliation fails

```
kubectl create -f control-plane/error-signalling/trigger-not-ready.yaml
```

## [Test] Trigger failure is propagated to the Ready status condition

The failure caused by the intentionally invalid spec is propagated to the `Ready` condition, which has a `status` of
`False`, a severity of `""` (error), and a non-empty `reason`:

```
kubectl get triggers/not-ready -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}{";"}{.status.conditions[?(@.type=="Ready")].severity}{";"}{.status.conditions[?(@.type=="Ready")].reason}'
```

### [Output]

```
False;;BrokerDoesNotExist
```

---

## [Pre] Create a Channel instance which reconciliation fails

```
kubectl create -f control-plane/error-signalling/channel-not-ready.yaml
```

## [Test] Channel failure is propagated to the Ready status condition

The failure caused by the intentionally invalid spec is propagated to the `Ready` condition, which has a `status` of
`Unknown` and a severity of `""` (error).

```
kubectl get channels/not-ready -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}{";"}{.status.conditions[?(@.type=="Ready")].severity}'
```

### [Output]

```
Unknown;
```

---

## [Pre] Create a Subscription instance which reconciliation fails

```
kubectl create -f control-plane/error-signalling/subscription-not-ready.yaml
```

## [Test] Subscription failure is propagated to the Ready status condition

The failure caused by the intentionally invalid spec is propagated to the `Ready` condition, which has a `status` of
`Unknown` and a severity of `""` (error).

```
kubectl get subscriptions/not-ready -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}{";"}{.status.conditions[?(@.type=="Ready")].severity}'
```

### [Output]

```
Unknown;
```
