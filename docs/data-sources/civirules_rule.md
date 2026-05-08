---
page_title: "civicrm_civirules_rule Data Source - CiviCRM"
subcategory: "CiviRules"
description: |-
  Fetches a CiviRules Rule by ID or name.
---

# civicrm_civirules_rule (Data Source)

Fetches a CiviRules Rule by ID or name. Use this data source to reference an existing rule without managing it in Terraform (e.g. to get its ID for cross-module references).

## Example Usage

```terraform
# Look up an existing rule by name
data "civicrm_civirules_rule" "existing" {
  name = "housing_income_stab_to_secured"
}

output "rule_id" {
  value = data.civicrm_civirules_rule.existing.id
}
```

## Argument Reference

At least one of `id` or `name` must be specified.

- `id` (Number, Optional) Unique ID of the rule.
- `name` (String, Optional) Machine name of the rule.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `label` (String) Human-readable label of the rule.
- `trigger_id` (Number) ID of the trigger that fires this rule.
- `trigger_params` (String) Trigger parameters as JSON string.
- `description` (String) Description of the rule.
- `help_text` (String) Help text shown in the CiviRules UI.
- `is_active` (Boolean) Whether the rule is active.
- `is_debug` (Boolean) Whether debug mode is enabled for this rule.
