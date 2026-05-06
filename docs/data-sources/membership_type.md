---
page_title: "civicrm_membership_type Data Source - CiviCRM"
subcategory: ""
description: |-
  Fetches a CiviCRM Membership Type by ID or name.
---

# civicrm_membership_type (Data Source)

Fetches a CiviCRM Membership Type by ID or name. Use this data source to look up existing membership types to reference in your configuration.

## Example Usage

```terraform
# Look up a membership type by name
data "civicrm_membership_type" "annual" {
  name = "Annual Membership"
}

# Look up a membership type by ID
data "civicrm_membership_type" "general" {
  id = 3
}

# Reference the data source in another resource
output "annual_fee" {
  value = data.civicrm_membership_type.annual.minimum_fee
}
```

## Argument Reference

The following arguments are supported. At least one of `id` or `name` must be specified.

- `id` (Number, Optional) The unique identifier of the membership type.
- `name` (String, Optional) The name of the membership type.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

- `description` (String) A description of the membership type.
- `member_of_contact_id` (Number) The contact ID of the owning organisation.
- `financial_type_id` (Number) The financial type ID used for membership fees.
- `minimum_fee` (Number) The minimum membership fee.
- `duration_unit` (String) The duration unit (`day`, `month`, `year`, `lifetime`).
- `duration_interval` (Number) The number of duration units.
- `period_type` (String) The period type (`rolling` or `fixed`).
- `fixed_period_start_day` (Number) The fixed period start day in `MMDD` format.
- `fixed_period_rollover_day` (Number) The fixed period rollover day in `MMDD` format.
- `relationship_type_id` (Number) The relationship type ID for inherited memberships.
- `relationship_direction` (String) The relationship direction for inherited memberships.
- `max_related` (Number) The maximum number of related memberships.
- `visibility` (String) Visibility setting (`Public` or `Admin`).
- `weight` (Number) The sort weight.
- `receipt_text_signup` (String) Text appended to signup receipts.
- `receipt_text_renewal` (String) Text appended to renewal receipts.
- `auto_renew` (Boolean) Whether auto-renewal is enabled.
- `is_active` (Boolean) Whether the membership type is active.
