resource "civicrm_civirules_rule_action" "set_status" {
  rule_id   = civicrm_civirules_rule.my_rule.id
  action_id = data.civicrm_civirules_action.change_case_status.id
  action_params = jsonencode({
    case_status_id = data.civicrm_case_status.resolved.value
  })
  is_active = true
}
