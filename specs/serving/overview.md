# Resource Types

The primary resources in the Knative Serving API are Routes, Revisions,
Configurations, and Services:

- A **Route** provides a named endpoint and a mechanism for routing traffic to

- **Revisions**, which are immutable snapshots of code + config, created by a Configuration.

- **Configuration**, which acts as a stream of environments for Revisions.

- A **Service** acts as a top-level container for managing a Route and
  Configuration which implement a network service.

![Object model](images/object_model.png)

## Route

A **Route** provides a network endpoint for a user's service (which consists of a
series of software and configuration Revisions over time). A Kubernetes
namespace can have multiple routes. The Route provides a long-lived, stable,
named, HTTP-addressable endpoint that is backed by one or more **Revisions**.
The default configuration is for the Route to automatically route traffic to the
latest Revision created by a **Configuration**. For more complex scenarios, the
API supports splitting traffic on a percentage basis, and CI tools could
maintain multiple configurations for a single Route (e.g. "golden path" and
“experiments”) or reference multiple Revisions directly to pin revisions during
an incremental rollout and n-way traffic split. The Route can optionally assign
addressable subdomains to any or all backing Revisions.

## Revision

A **Revision** is an immutable snapshot of code and configuration. A Revision
references a container image. Revisions are created by updates to a
**Configuration**.

Revisions that are not addressable via a Route may be garbage collected and all
underlying Kubernetes resources will be deleted. Revisions that are addressable via a
Route will have resource utilization proportional to the load they are under.

## Configuration

A **Configuration** describes the desired latest Revision state, and creates and
tracks the status of Revisions as the desired state is updated. A configuration
will reference a container image and associated execution metadata needed by the
Revision. On updates to a Configuration's spec, a new Revision will be created;
the Configuration's controller will track the status of created Revisions and
makes the most recently created and most recently _ready_ Revisions available in
the status section.

## Service

A **Service** encapsulates a **Route** and **Configuration** which together
provide a software component. A Service exists to provide a singular abstraction
which can be access controlled, reasoned about, and which encapsulates software
lifecycle decisions such as rollout policy and team resource ownership. A Service
acts only as an orchestrator of the underlying Route and Configuration (much as
a Kubernetes Deployment orchestrates ReplicaSets). Its usage is optional but
recommended.

The Service's controller will track the statuses of its owned Configuration and
Route, reflecting their statuses and conditions as its own.

The owned Configuration's Ready conditions are surfaced as the Service's
ConfigurationsReady condition. The owned Routes' Ready conditions are surfaced
as the Service's RoutesReady condition.

## Orchestration

Revisions are created indirectly when a Configuration is created or updated.
This provides:

- a single referenceable resource for the Route to perform automated rollouts
- a single resource that can be watched to see a history of all the Revisions
  created
- PATCH semantics for Revisions implemented server-side, minimizing
  read-modify-write implemented across multiple clients, which could result in
  optimistic concurrency errors
- the ability to rollback to a known good Configuration

Update operations on the Service enable scenarios such as:

- _"Push image, keep config":_ Specifying a new Revision with updated image,
  inheriting configuration such as env vars from the Configuration
- _"Update config, keep image"_: Specifying a new Revision as just a change to
  Configuration, such as updating an env variable, inheriting all other
  configuration and image
- _"Execute a controlled rollout"_: Updating the Service's traffic spec allows
  testing of Revisions before making them live, and controlled rollouts
