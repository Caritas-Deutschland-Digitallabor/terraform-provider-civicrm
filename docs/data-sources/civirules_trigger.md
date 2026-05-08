---
page_title: "civicrm_civirules_trigger Data Source - CiviCRM"
subcategory: "CiviRules"
description: |-
  Fetches a CiviRules Trigger by ID or name.
---

# civicrm_civirules_trigger (Data Source)

Fetches a CiviRules Trigger by ID or name. Use this data source to look up the `trigger_id` for [`civicrm_civirules_rule`](../resources/civirules_rule.md) without hardcoding numeric IDs.

## Example Usage

```terraform
# Look up by machine name
data "civicrm_civirules_trigger" "activity_changed" {
  name = "activity_is_changed"
}

data "civicrm_civirules_trigger" "case_changed" {
  name = "case_is_changed"
}

resource "civicrm_civirules_rule" "my_rule" {
  name       = "my_automation_rule"
  label      = "My Automation"
  trigger_id = data.civicrm_civirules_trigger.activity_changed.id
  is_active  = true
}
```

## Argument Reference

At least one of `id` or `name` must be specified.

- `id` (Number, Optional) Unique ID of the trigger.
- `name` (String, Optional) Machine name of the trigger (e.g. `"activity_is_changed"`, `"case_is_changed"`).

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `label` (String) Display label shown in the CiviRules rule form.
- `object_name` (String) CiviCRM entity this trigger fires on.
- `op` (String) Operation: `create`, `edit`, or `delete`.
- `cron` (Boolean) Whether this is a cron trigger.
- `class_name` (String) PHP class implementing the trigger.
- `is_active` (Boolean) Whether the trigger is active.
