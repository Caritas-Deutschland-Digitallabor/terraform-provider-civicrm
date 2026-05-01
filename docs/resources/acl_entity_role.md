---
page_title: "civicrm_acl_entity_role Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM ACL Entity Role assignment. This assigns ACL roles to groups of contacts.
---

# civicrm_acl_entity_role (Resource)

Manages a CiviCRM ACL Entity Role assignment. This resource connects ACL roles to groups, granting all members of the group the permissions defined by the role's ACL rules.

## Example Usage

```terraform
# Create an ACL role
resource "civicrm_acl_role" "volunteer_manager" {
  name        = "volunteer_manager"
  label       = "Volunteer Manager"
  description = "Can manage volunteer contacts"
  is_active   = true
}

# Create a group for team leaders
resource "civicrm_group" "team_leaders" {
  name        = "team_leaders"
  title       = "Team Leaders"
  description = "Staff members with team lead responsibilities"
  is_active   = true
  group_type  = ["Access Control"]
}

# Assign the ACL role to the group
# Note: acl_role_id must reference the ACL role's 'value' field, not 'id'
resource "civicrm_acl_entity_role" "team_leaders_as_managers" {
  acl_role_id  = tonumber(civicrm_acl_role.volunteer_manager.value)
  entity_table = "civicrm_group"
  entity_id    = civicrm_group.team_leaders.id
  is_active    = true
}
```

~> **Important:** The `acl_role_id` field must reference the ACL role's `value` attribute, not its `id`. CiviCRM internally uses the `value` field to link role assignments. Use `tonumber()` to convert the string value to a number.

## Argument Reference

The following arguments are supported:

### Required

- `acl_role_id` (Number) The `value` field of the ACL role to assign. Use `tonumber(civicrm_acl_role.example.value)` to reference an ACL role.
- `entity_id` (Number) The ID of the entity (e.g., group) to assign the role to.
- `entity_table` (String) The entity type. Currently only `civicrm_group` is supported.

### Optional

- `is_active` (Boolean) Whether this assignment is active. Default: `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the ACL entity role assignment.

## Import

ACL Entity Roles can be imported using the assignment ID:

```shell
terraform import civicrm_acl_entity_role.example 123
```
