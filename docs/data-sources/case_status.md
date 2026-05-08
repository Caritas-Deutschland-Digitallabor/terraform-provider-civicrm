---
page_title: "civicrm_case_status Data Source - CiviCRM"
subcategory: "Cases"
description: |-
  Fetches a CiviCRM Case Status by ID or name.
---

# civicrm_case_status (Data Source)

Fetches a CiviCRM Case Status (an `OptionValue` in the `case_status` option group) by ID or name. Particularly useful to look up the `value` field for use in CiviRules `action_params` when changing case status.

## Example Usage

```terraform
data "civicrm_case_status" "housing_secured" {
  name = "housing_secured"
}

data "civicrm_case_status" "case_closed" {
  name = "Closed"
}

# Use value in a CiviRules action
resource "civicrm_civirules_rule_action" "close_case" {
  rule_id   = civicrm_civirules_rule.my_rule.id
  action_id = data.civicrm_civirules_action.change_case_status.id
  action_params = jsonencode({
    case_status_id = data.civicrm_case_status.case_closed.value
  })
  is_active = true
}
```

## Argument Reference

At least one of `id` or `name` must be specified.

- `id` (Number, Optional) Unique ID of the OptionValue record.
- `name` (String, Optional) Machine name of the case status.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `label` (String) Display label shown in the UI.
- `grouping` (String) Category: `"Opened"` for active cases, `"Closed"` for resolved cases.
- `is_active` (Boolean) Whether the status is active.
- `is_reserved` (Boolean) Whether the status is reserved by the system.
- `weight` (Number) Sort weight.
- `value` (String) Internal value used by CiviCRM. Reference this in CiviRules `action_params` for status changes.
