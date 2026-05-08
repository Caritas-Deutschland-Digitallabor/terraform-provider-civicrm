---
page_title: "civicrm_civirules_action Data Source - CiviCRM"
subcategory: "CiviRules"
description: |-
  Fetches a CiviRules Action type by ID or name.
---

# civicrm_civirules_action (Data Source)

Fetches a CiviRules Action type by ID or name. Use the returned `id` as the `action_id` in [`civicrm_civirules_rule_action`](../resources/civirules_rule_action.md).

## Example Usage

```terraform
data "civicrm_civirules_action" "change_case_status" {
  name = "change_case_status"
}

data "civicrm_civirules_action" "send_email" {
  name = "send_email"
}

resource "civicrm_civirules_rule_action" "set_status" {
  rule_id   = civicrm_civirules_rule.my_rule.id
  action_id = data.civicrm_civirules_action.change_case_status.id
  action_params = jsonencode({
    case_status_id = data.civicrm_case_status.resolved.value
  })
  is_active = true
}
```

## Argument Reference

At least one of `id` or `name` must be specified.

- `id` (Number, Optional) Unique ID of the action type.
- `name` (String, Optional) Machine name of the action type (e.g. `"change_case_status"`, `"send_email"`).

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `label` (String) Display label shown in the CiviRules UI.
- `class_name` (String) PHP class implementing the action.
- `is_active` (Boolean) Whether the action type is active.
