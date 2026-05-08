resource "civicrm_civirules_rule_condition" "only_follow_ups" {
  rule_id      = civicrm_civirules_rule.my_rule.id
  condition_id = data.civicrm_civirules_condition.activity_type.id
  condition_params = jsonencode({
    activity_type_id = "Follow up"
  })
  is_active = true
  negate    = false
}
