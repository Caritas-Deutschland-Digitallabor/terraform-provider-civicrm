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
