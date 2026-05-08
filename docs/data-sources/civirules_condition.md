---
page_title: "civicrm_civirules_condition Data Source - CiviCRM"
subcategory: "CiviRules"
description: |-
  Fetches a CiviRules Condition type by ID or name.
---

# civicrm_civirules_condition (Data Source)

Fetches a CiviRules Condition type by ID or name. Use the returned `id` as the `condition_id` in [`civicrm_civirules_rule_condition`](../resources/civirules_rule_condition.md).

## Example Usage

```terraform
data "civicrm_civirules_condition" "activity_type" {
  name = "activity_type"
}

data "civicrm_civirules_condition" "activity_status" {
  name = "activity_status"
}

resource "civicrm_civirules_rule_condition" "only_follow_ups" {
  rule_id      = civicrm_civirules_rule.my_rule.id
  condition_id = data.civicrm_civirules_condition.activity_type.id
  condition_params = jsonencode({
    activity_type_id = "Follow up"
  })
  is_active = true
}
```

## Argument Reference

At least one of `id` or `name` must be specified.

- `id` (Number, Optional) Unique ID of the condition type.
- `name` (String, Optional) Machine name of the condition type (e.g. `"activity_type"`, `"case_type_is"`).

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `label` (String) Display label shown in the CiviRules UI.
- `class_name` (String) PHP class implementing the condition.
- `is_active` (Boolean) Whether the condition type is active.
