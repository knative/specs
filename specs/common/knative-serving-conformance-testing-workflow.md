# Knative Conformance Workflow

This document covers the workflow steps to gain knative conformance

## Conformance Steps
1. Run Manual Tests
2. Open Github PR in knative/specs repo 
  - Use PR template below
  - place in the following directory structure
    - conformance/tests/results/$spec_version/${date}-${product} directory with a metadata file and test results
    - ex; .../v1.0/2021-12-18-ibm-code-engine/...
  - Attach Logs from tests
3. KTC reviews 
  - Template
  - logs
4. Knative Trademark Committee (KTC) provide approval / exception / or non-approval
  - The KTC can be contacted at trademark@knative.team 
6. Post to conformance matrix - consumable by the general public as a table of version, product, vendor and outcomes.
  - (TBD) Should have a public “conformant implementations” showcase on https://knative.dev/offerings



## Issue Template

| **Title** | *Fill out with* |
| ------------------ | -------------------------------------------- |
| **Vendor** | *Name of Company / Entity* |
| **Product Name** | *Product Name* |
| **Version / Version** | *Knative version / Knative conformance ver* |
| **Website URL for Product** | *Website of Product* |
| **Contact Name** | *Conformance contact for questions* |




