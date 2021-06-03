# Error Signalling

<!-- copied from ../serving/knative-api-specification-1.0.md#error-signalling -->

Knative APIs use the
[Kubernetes Conditions convention](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties)
to communicate errors and problems to the user. Note that Knative customizes the
general Kubernetes recommendation with a `severity` field, and does not include
`lastTransitionTime` for scalability reasons. Each user-visible resource
described in Resource Overview MUST have a `conditions` field in `status`, which
MUST be a list of `Condition` objects of the following form. Fields in the
condition which are not marked as REQUIRED may be omitted to indicate the
default value (i.e. a Condition which does not include a `status` field is
equivalent to a `status` of `"Unknown"`). The actual API object types in an
OpenAPI document may be named `FooCondition` to allow better code generation and
disambiguation between similar fields in the same `apiGroup`.

<table>
  <tr>
   <td><strong>Field</strong>
   </td>
   <td><strong>Type</strong>
   </td>
   <td><strong>Description</strong>
   </td>
   <td><strong>Default Value If Unset</strong>
   </td>
  </tr>
  <tr>
   <td><code>type</code>
   </td>
   <td><code>string</code>
   </td>
   <td>The category of the condition, as a short, CamelCase word or phrase.
<p>
This is the primary key of the Conditions list when viewed as a map.
   </td>
   <td>REQUIRED – No default
   </td>
  </tr>
  <tr>
   <td><code>status</code>
   </td>
   <td>Enum:<ul>

<li>"True"
<li>"False"
<li>"Unknown"</li></ul>

   </td>
   <td>The last measured status of this condition.
   </td>
   <td>"Unknown"
   </td>
  </tr>
  <tr>
   <td><code>reason</code>
   </td>
   <td>string
   </td>
   <td>One-word CamelCase reason for the condition's last transition.
   </td>
   <td>""
   </td>
  </tr>
  <tr>
   <td><code>message</code>
   </td>
   <td>string
   </td>
   <td>Human-readable sentence describing the last transition.
   </td>
   <td>""
   </td>
  </tr>
  <tr>
   <td><code>severity</code>
   </td>
   <td>Enum:<ul>

<li>""
<li>"Warning"
<li>"Info"</li></ul>

   </td>
   <td>If present, represents the severity of the condition. An empty severity represents a severity level of "Error". 
   </td>
   <td>""
   </td>
  </tr>
  <tr>
   <td><code>lastTransitionTime</code>
   </td>
   <td>Timestamp
   </td>
   <td>Last update time for this condition.
   </td>
   <td>(no timestamp specified)
   </td>
  </tr>
</table>

Additionally, the resource's `status.conditions` field MUST be managed as
follows to enable clients (particularly user interfaces) to present useful
diagnostic and error message to the user. In the following section, conditions
are referred to by their `type` (aka the string value of the `type` field on the
Condition).

1.  Each resource MUST have either a `Ready` condition (for ongoing systems) or
    `Succeeded` condition (for resources that run to completion) with
    `severity=""`, which MUST use the `"True"`, `"False"`, and `"Unknown"`
    status values as follows:

    1.  `"False"` MUST indicate a failure condition.
    1.  `"Unknown"` SHOULD indicate that reconciliation is not yet complete and
        success or failure is not yet determined.
    1.  `"True"` SHOULD indicate that the resource is fully reconciled and
        operating correctly.

    `"Unknown"` and `"True"` are specified as SHOULD rather than MUST
    requirements because there may be errors which prevent functioning which
    cannot be determined by the API stack (e.g. DNS record configuration in
    certain environments). Implementations are expected to treat these as "MUST"
    for factors within the control of the implementation.

1.  For non-`Ready` conditions, any conditions with `severity=""` (aka "Error
    conditions") MUST be aggregated into the "Ready" condition as follows:

    1.  If the condition is `"False"`, `Ready`'s status MUST be `"False"`.
    1.  If the condition is `"Unknown"`, `Ready`'s status MUST be `"False"` or
        `"Unknown"`.
    1.  If the condition is `"True"`, `Ready`'s status can be any of `"True"`,
        `"False"`, or `"Unknown"`.

    Implementations MAY choose to report that `Ready` is `"False"` or
    `"Unknown"` even if all Error conditions report a status of `"True"` (i.e.
    there might be additional hidden implementation conditions which feed into
    the `Ready` condition which are not reported.)

1.  Non-`Ready` conditions with non-error severity MAY be surfaced by the
    implementation. Examples of `Warning` or `Info` conditions could include:
    missing health check definitions, scale-to-zero status, or non-fatal
    capacity limits.

1.  Conditions with a `status` other than `"True"` SHOULD provide `message` and
    `reason` fields indicating the reason that the `status` is not `"True"`.
    Conditions where the `status` is `"False"` MUST provide a failure `reason`
    in the condition. (`"Unknown"` conditions may not have been reconciled, and
    so may have an empty `reason`.)

Conditions type names SHOULD be chosen to describe positive conditions where
`"True"` means that the condition has been satisfied. Some conditions MAY be
transient (for example, `ResourcesAllocated` might change between `"True"` and
`"False"` as an application scales to and from zero). It is RECOMMENDED that
transient conditions be indicated with a `severity="Info"`.
