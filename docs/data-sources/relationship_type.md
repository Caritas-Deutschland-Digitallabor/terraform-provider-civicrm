---
page_title: "civicrm_relationship_type Data Source - CiviCRM"
subcategory: ""
description: |-
  Fetches a CiviCRM Relationship Type by ID or name.
---

# civicrm_relationship_type (Data Source)

Fetches a CiviCRM Relationship Type by ID or `name_a_b`. Use this to reference existing relationship types (e.g. case roles) without managing them via Terraform.

## Example Usage

```terraform
# Look up by machine name (A→B direction)
data "civicrm_relationship_type" "housing_coordinator" {
  name_a_b = "Housing Coordinator"
}

# Look up by ID
data "civicrm_relationship_type" "employee" {
  id = 5
}

output "coordinator_label" {
  value = data.civicrm_relationship_type.housing_coordinator.label_a_b
}
```

## Argument Reference

At least one of `id` or `name_a_b` must be specified.

- `id` (Number, Optional) Unique ID of the relationship type.
- `name_a_b` (String, Optional) Machine name of the relationship in the A→B direction (e.g. `"Housing Coordinator"`).

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `label_a_b` (String) Display label in the A→B direction.
- `name_b_a` (String) Machine name in the B→A direction.
- `label_b_a` (String) Display label in the B→A direction.
- `description` (String) Description of the relationship type.
- `contact_type_a` (String) Required contact type for side A, or empty for any.
- `contact_type_b` (String) Required contact type for side B, or empty for any.
- `contact_sub_type_a` (String) Required contact subtype for side A.
- `contact_sub_type_b` (String) Required contact subtype for side B.
- `is_reserved` (Boolean) Whether this is a system-reserved type.
- `is_active` (Boolean) Whether the relationship type is active.
