---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "spire_entry Data Source - spire"
subcategory: ""
description: |-
  Register workloads with Spire Server Entry
---

# spire_entry (Data Source)

Register workloads with Spire Server Entry



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `spiffe_id` (Object) The SPIFFE ID of this record's parent (see [below for nested schema](#nestedatt--spiffe_id))

### Read-Only

- `id` (String) The Registration Entry ID of the record
- `parent_id` (Object) The SPIFFE ID of this record's parent (see [below for nested schema](#nestedatt--parent_id))
- `selectors` (Attributes Set) A type/value selector. Can be used more than once (see [below for nested schema](#nestedatt--selectors))

<a id="nestedatt--spiffe_id"></a>
### Nested Schema for `spiffe_id`

Required:

- `path` (String)
- `trust_domain` (String)


<a id="nestedatt--parent_id"></a>
### Nested Schema for `parent_id`

Read-Only:

- `path` (String)
- `trust_domain` (String)


<a id="nestedatt--selectors"></a>
### Nested Schema for `selectors`

Required:

- `type` (String)
- `value` (String)