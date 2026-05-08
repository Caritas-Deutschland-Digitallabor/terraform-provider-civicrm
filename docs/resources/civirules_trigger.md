---
page_title: "civicrm_civirules_trigger Resource - CiviCRM"
subcategory: "CiviRules"
description: |-
  Manages a CiviRules Trigger definition.
---

# civicrm_civirules_trigger (Resource)

Manages a CiviRules Trigger definition (entity: `CiviRulesTrigger`). Most standard triggers ("Case is changed", "Activity is added", etc.) are shipped by the CiviRules extension and should be referenced by ID in [`civicrm_civirules_rule`](civirules_rule.md) rather than created here. Use this resource only for **custom cron triggers** or triggers backed by a custom PHP class.

To look up an existing trigger's ID, use the [`civicrm_civirules_trigger`](../data-sources/civirules_trigger.md) data source.

## Example Usage

```terraform
# Custom cron trigger (runs on CiviCRM's scheduled jobs)
resource "civicrm_civirules_trigger" "weekly_check" {
  name       = "my_weekly_housing_check"
  label      = "Weekly Housing Status Check"
  cron       = true
  class_name = "CRM_MyExtension_Trigger_WeeklyCheck"
  is_active  = true
}

# Custom event-based trigger for a specific entity operation
resource "civicrm_civirules_trigger" "custom_case_created" {
  name        = "my_case_created_trigger"
  label       = "Custom: Case Created"
  object_name = "Case"
  op          = "create"
  class_name  = "CRM_MyExtension_Trigger_CaseCreated"
  is_active   = true
}
```

## Argument Reference

The following arguments are supported:

### Required

- `name` (String) Machine name of the trigger. Must be unique.
- `label` (String) Human-readable label shown in the CiviRules rule form.

### Optional

- `object_name` (String) CiviCRM entity this trigger fires on (e.g. `"Case"`, `"Activity"`, `"Contact"`). Leave empty for cron triggers.
- `op` (String) Operation that fires the trigger: `"create"`, `"edit"`, `"delete"`, or empty for cron triggers.
- `cron` (Boolean) Whether this is a cron-based (scheduled) trigger. Default: `false`.
- `class_name` (String) Fully-qualified PHP class name that implements the trigger logic (e.g. `"CRM_MyExt_Trigger_MyTrigger"`).
- `is_active` (Boolean) Whether this trigger is active. Default: `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the trigger.

## Import

CiviRules Triggers can be imported using the trigger ID:

```shell
terraform import civicrm_civirules_trigger.example 1
```
