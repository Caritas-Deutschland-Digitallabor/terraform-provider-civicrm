---
page_title: "civicrm_acl Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM ACL rule. ACL rules define what operations an ACL role can perform on specific objects.
---

# civicrm_acl (Resource)

Manages a CiviCRM ACL rule. ACL rules define what operations (View, Edit, Create, Delete, Search) an ACL role can perform on specific objects like groups.

## Example Usage

```terraform
# Grant edit access to a group
# Note: entity_id must reference the ACL role's 'value' field, not 'id'
resource "civicrm_acl" "managers_edit_volunteers" {
  name         = "managers_edit_volunteers"
  entity_id    = tonumber(civicrm_acl_role.volunteer_manager.value)
  operation    = "Edit"
  object_table = "civicrm_group"
  object_id    = civicrm_group.volunteers.id
  is_active    = true
}

# Grant view-only access
resource "civicrm_acl" "viewers_view_volunteers" {
  name         = "viewers_view_volunteers"
  entity_id    = tonumber(civicrm_acl_role.data_viewer.value)
  operation    = "View"
  object_table = "civicrm_group"
  object_id    = civicrm_group.volunteers.id
  is_active    = true
}

# Grant access to all contacts (object_id = 0)
resource "civicrm_acl" "admins_edit_all" {
  name         = "admins_edit_all_contacts"
  entity_id    = tonumber(civicrm_acl_role.admin.value)
  operation    = "Edit"
  object_table = "civicrm_group"
  object_id    = 0
  is_active    = true
}
```

~> **Important:** The `entity_id` field must reference the ACL role's `value` attribute, not its `id`. CiviCRM internally uses the `value` field to link ACL rules to roles. Use `tonumber()` to convert the string value to a number.

## Argument Reference

The following arguments are supported:

### Required

- `entity_id` (Number) The `value` field of the ACL role this rule applies to. Use `tonumber(civicrm_acl_role.example.value)` to reference an ACL role.
- `name` (String) The machine name of the ACL rule (must be unique).
- `object_id` (Number) The ID of the object (e.g., group ID) this rule applies to. Use `0` for all objects.
- `object_table` (String) The table/entity type this rule applies to (e.g., `civicrm_group`).
- `operation` (String) The operation this rule permits. Valid values: `View`, `Edit`, `Create`, `Delete`, `Search`, `All`.

### Optional

- `deny` (Boolean) Whether this rule denies (rather than grants) the operation. Default: `false`.
- `is_active` (Boolean) Whether this ACL rule is active. Default: `true`.
- `priority` (Number) The priority of this rule (higher numbers take precedence). Default: `0`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the ACL rule.

## Import

ACL rules can be imported using the rule ID:

```shell
terraform import civicrm_acl.example 123
```
