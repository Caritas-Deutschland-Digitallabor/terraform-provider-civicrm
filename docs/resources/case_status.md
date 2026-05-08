---
page_title: "civicrm_case_status Resource - CiviCRM"
subcategory: "Cases"
description: |-
  Manages a CiviCRM Case Status.
---

# civicrm_case_status (Resource)

Manages a CiviCRM Case Status. Case statuses are `OptionValue` records in the `case_status` option group. They control which lifecycle stage a case is in and how it appears in reports, dashboards, and searches.

The `grouping` field (`"Opened"` or `"Closed"`) determines whether CiviCRM treats the case as active or resolved. The `value` field is the internal identifier used in CiviRules `action_params` when changing case status.

## Example Usage

```terraform
resource "civicrm_case_status" "housing_secured" {
  name      = "housing_secured"
  label     = "Housing Secured"
  grouping  = "Opened"
  is_active = true
  weight    = 10
}

resource "civicrm_case_status" "case_closed" {
  name      = "case_closed"
  label     = "Case Closed"
  grouping  = "Closed"
  is_active = true
  weight    = 90
}

# Use the value in a CiviRules action
resource "civicrm_civirules_rule_action" "set_status" {
  rule_id   = civicrm_civirules_rule.my_rule.id
  action_id = data.civicrm_civirules_action.change_case_status.id
  action_params = jsonencode({
    case_status_id = civicrm_case_status.housing_secured.value
  })
  is_active = true
}
```

## Argument Reference

The following arguments are supported:

### Required

- `name` (String) Machine name of the status (e.g. `"housing_secured"`). Must be unique within the `case_status` option group.
- `label` (String) Display label shown in the UI (e.g. `"Housing Secured"`).

### Optional

- `grouping` (String) Category bucket for the status. Use `"Opened"` for active/open cases and `"Closed"` for resolved/ended cases. CiviCRM uses this to filter cases in dashboards and reports.
- `is_active` (Boolean) Whether the status is active. Default: `true`.
- `is_reserved` (Boolean) Whether the status is reserved (protected from deletion by CiviCRM). Default: `false`.
- `weight` (Number) Sort weight. Controls display order in dropdowns.
- `value` (String) Internal value used by CiviCRM to identify this status. Auto-generated if not set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the OptionValue record.
- `value` (String) The internal value assigned by CiviCRM (computed if not specified). Reference this in CiviRules `action_params` for status changes.
- `weight` (Number) The sort weight (computed if not specified).
- `grouping` (String) The grouping value (computed if not specified).

## Import

Case Statuses can be imported using the OptionValue ID:

```shell
terraform import civicrm_case_status.example 42
```
