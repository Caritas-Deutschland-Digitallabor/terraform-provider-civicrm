# Assign the ACL role to a group
# Note: acl_role_id must reference the ACL role's 'value' field, not 'id'
resource "civicrm_acl_entity_role" "team_leaders_as_managers" {
  acl_role_id  = tonumber(civicrm_acl_role.volunteer_manager.value)
  entity_table = "civicrm_group"
  entity_id    = civicrm_group.team_leaders.id
  is_active    = true
}
