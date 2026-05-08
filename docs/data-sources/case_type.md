---
page_title: "civicrm_case_type Data Source - CiviCRM"
subcategory: "Cases"
description: |-
  Fetches a CiviCRM Case Type by ID or name.
---

# civicrm_case_type (Data Source)

Fetches a CiviCRM Case Type by ID or name. Use this data source to reference an existing case type (e.g. to pass its ID into CiviRules `trigger_params`) without managing it in Terraform.

## Example Usage

```terraform
data "civicrm_case_type" "housing_support" {
  name = "housing_support"
}

# Reference the case type ID in a CiviRules rule trigger
resource "civicrm_civirules_rule" "housing_rule" {
  name       = "housing_case_activity_rule"
  label      = "Housing: Activity automation"
  trigger_id = data.civicrm_civirules_trigger.case_activity_changed.id
  trigger_params = jsonencode({
    case_type_id = data.civicrm_case_type.housing_support.id
  })
  is_active = true
}

output "housing_support_id" {
  value = data.civicrm_case_type.housing_support.id
}
```

## Argument Reference

At least one of `id` or `name` must be specified.

- `id` (Number, Optional) Unique ID of the case type.
- `name` (String, Optional) Machine name of the case type.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `title` (String) Display title of the case type.
- `description` (String) Description of the case type.
- `is_active` (Boolean) Whether the case type is active.
- `is_reserved` (Boolean) Whether the case type is reserved by the system.
- `weight` (Number) Sort weight.
- `definition` (String) Case type definition as a JSON string (contains `activityTypes`, `activitySets`, `caseRoles`).
