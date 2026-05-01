---
page_title: "civicrm_acl_role Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM ACL Role. ACL Roles define permission sets that can be assigned to groups.
---

# civicrm_acl_role (Resource)

Manages a CiviCRM ACL Role. ACL Roles define permission sets that can be assigned to groups of contacts through ACL Entity Roles.

## Example Usage

```terraform
# Basic ACL role (CiviCRM auto-generates the value)
resource "civicrm_acl_role" "volunteer_manager" {
  name        = "volunteer_manager"
  label       = "Volunteer Manager"
  description = "Can view and edit volunteer contacts"
  is_active   = true
}

# ACL role with explicit value for predictable references
resource "civicrm_acl_role" "data_viewer" {
  name        = "data_viewer"
  label       = "Data Viewer"
  description = "Read-only access to contact data"
  value       = "100"
  is_active   = true
}

# Combined with ACL rules - use tonumber() to convert value to number
resource "civicrm_acl_role" "event_coordinator" {
  name        = "event_coordinator"
  label       = "Event Coordinator"
  description = "Manages event participants"
  value       = "101"
  is_active   = true
}

resource "civicrm_acl" "event_coordinator_edit" {
  name         = "event_coordinator_edit_participants"
  entity_id    = tonumber(civicrm_acl_role.event_coordinator.value)
  operation    = "Edit"
  object_table = "civicrm_group"
  object_id    = civicrm_group.event_participants.id
  is_active    = true
}
```

## Argument Reference

The following arguments are supported:

### Required

- `label` (String) The display label of the ACL role.
- `name` (String) The machine name of the ACL role (must be unique).

### Optional

- `description` (String) A description of the ACL role.
- `is_active` (Boolean) Whether the ACL role is active. Default: `true`.
- `value` (String) The internal value used by CiviCRM to link ACL rules and entity roles. If not specified, CiviCRM auto-generates it. Use `tonumber(civicrm_acl_role.example.value)` when referencing in `civicrm_acl` or `civicrm_acl_entity_role` resources.
- `weight` (Number) The sort weight of the ACL role.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the ACL role.
- `value` (String) The internal value (computed if not specified).

## Import

ACL Roles can be imported using the role ID:

```shell
terraform import civicrm_acl_role.example 123
```
