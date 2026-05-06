---
page_title: "civicrm_membership_type Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM Membership Type.
---

# civicrm_membership_type (Resource)

Manages a CiviCRM Membership Type. Membership Types define the rules for memberships, including fees, duration, renewal behaviour, and which organisation they belong to.

## Example Usage

```terraform
# Rolling annual membership
resource "civicrm_membership_type" "annual" {
  name                 = "Annual Membership"
  description          = "Standard annual membership, renews on join date"
  member_of_contact_id = 1
  financial_type_id    = 2
  minimum_fee          = 50.00
  duration_unit        = "year"
  duration_interval    = 1
  period_type          = "rolling"
  visibility           = "Public"
  is_active            = true
}

# Fixed-period membership (calendar year, Jan 1 – Dec 31, rollover Oct 1)
resource "civicrm_membership_type" "calendar_year" {
  name                    = "Calendar Year Membership"
  description             = "Membership runs from January 1 to December 31"
  member_of_contact_id    = 1
  financial_type_id       = 2
  minimum_fee             = 40.00
  duration_unit           = "year"
  duration_interval       = 1
  period_type             = "fixed"
  fixed_period_start_day  = 101   # January 1 (MMDD)
  fixed_period_rollover_day = 1001 # October 1 (MMDD)
  visibility              = "Public"
  is_active               = true
}

# Free lifetime membership
resource "civicrm_membership_type" "lifetime" {
  name                 = "Lifetime Membership"
  description          = "One-time free membership, never expires"
  member_of_contact_id = 1
  financial_type_id    = 2
  minimum_fee          = 0.00
  duration_unit        = "lifetime"
  period_type          = "rolling"
  visibility           = "Public"
  is_active            = true
}

# Inherited membership via a relationship
resource "civicrm_membership_type" "family" {
  name                   = "Family Membership"
  description            = "Extends to household members via relationship"
  member_of_contact_id   = 1
  financial_type_id      = 2
  minimum_fee            = 80.00
  duration_unit          = "year"
  duration_interval      = 1
  period_type            = "rolling"
  relationship_type_id   = 7
  relationship_direction = "b_a"
  max_related            = 4
  visibility             = "Public"
  is_active              = true
}
```

## Argument Reference

The following arguments are supported:

### Required

- `name` (String) The name of the membership type (must be unique within the organisation).
- `member_of_contact_id` (Number) The contact ID of the organisation this membership type belongs to.
- `financial_type_id` (Number) The financial type ID used for membership fee transactions.
- `duration_unit` (String) The unit of the membership duration. Valid values: `day`, `month`, `year`, `lifetime`.
- `period_type` (String) How the membership period is calculated. Valid values: `rolling` (starts on the join date), `fixed` (starts on a fixed calendar date).

### Optional

- `description` (String) A description of the membership type.
- `minimum_fee` (Number) The minimum membership fee. Default: `0`.
- `duration_interval` (Number) The number of `duration_unit`s the membership lasts. Not applicable for `lifetime`.
- `fixed_period_start_day` (Number) For `fixed` period memberships: the start day in `MMDD` format (e.g. `101` for January 1). Default: `101`.
- `fixed_period_rollover_day` (Number) For `fixed` period memberships: the rollover day in `MMDD` format. Memberships that join after this day roll into the next period.
- `relationship_type_id` (Number) The relationship type ID used for inherited memberships (a member's related contacts also become members).
- `relationship_direction` (String) The direction of the relationship for inherited memberships (e.g. `b_a`).
- `max_related` (Number) The maximum number of related memberships that can be inherited. `0` means unlimited.
- `visibility` (String) Whether the membership type is publicly visible. Valid values: `Public`, `Admin`. Default: `Public`.
- `weight` (Number) The sort weight controlling display order.
- `receipt_text_signup` (String) Additional text appended to the receipt email sent on signup.
- `receipt_text_renewal` (String) Additional text appended to the receipt email sent on renewal.
- `auto_renew` (Boolean) Whether this membership type supports automatic renewal. Default: `false`.
- `is_active` (Boolean) Whether the membership type is active. Default: `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the membership type.

## Understanding Period Types

**Rolling** (`period_type = "rolling"`): The membership starts on the date the contact joins and ends after the specified `duration_interval` × `duration_unit`. Each renewal extends from the current expiry date.

**Fixed** (`period_type = "fixed"`): All memberships of this type share the same start date regardless of when a contact joins. Use `fixed_period_start_day` to define that date. Use `fixed_period_rollover_day` to define the cutoff: contacts who join after the rollover day are counted as joining for the *next* fixed period instead.

## Inherited Memberships

Setting `relationship_type_id` enables membership inheritance: when contact A holds this membership, related contacts (linked via the specified relationship type and direction) automatically receive a related membership. `max_related` limits how many such related memberships are created.

## Import

Membership Types can be imported using the membership type ID:

```shell
terraform import civicrm_membership_type.example 5
```
