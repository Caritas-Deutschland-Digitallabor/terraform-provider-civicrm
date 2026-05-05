# ─────────────────────────────────────────────────────────────────────────────
# CiviRules: Housing Support Case Automation
#
# This example shows a complete rule that:
#   WHEN  an activity is completed on a housing-support case
#   IF    the activity is "Income and benefits stabilization"
#   THEN  change the case status to "Housing Secured"
#
# IMPORTANT: trigger_id, condition_id, action_id are numeric IDs that depend on
# your CiviRules installation. Look them up with:
#   CiviRulesTrigger.get  → find the "Activity is changed" trigger
#   CiviRulesCondition.get → find conditions like "Activity type is"
#   CiviRulesAction.get   → find actions like "Change case status"
# ─────────────────────────────────────────────────────────────────────────────

resource "civicrm_civirules_rule" "income_stab_closes_case" {
  name      = "housing_income_stab_to_secured"
  label     = "Housing: Income stabilization → Status Housing Secured"
  # trigger_id: ID of "Activity is changed" trigger from CiviRulesTrigger.get
  trigger_id = 5
  description = "When the income stabilization activity is completed, mark the case as Housing Secured."
  is_active  = true
}

# Condition: activity type must be "Income and benefits stabilization"
# condition_id: ID of "Activity type is" condition from CiviRulesCondition.get
resource "civicrm_civirules_rule_condition" "check_activity_type" {
  rule_id      = civicrm_civirules_rule.income_stab_closes_case.id
  condition_id = 3
  condition_params = jsonencode({
    activity_type_id = "Income and benefits stabilization"
  })
  is_active = true
  negate    = false
}

# Condition: activity status must be "Completed"
# condition_id: ID of "Activity status is" condition from CiviRulesCondition.get
resource "civicrm_civirules_rule_condition" "check_activity_status" {
  rule_id      = civicrm_civirules_rule.income_stab_closes_case.id
  condition_id = 4
  condition_params = jsonencode({
    activity_status_id = "2" # 2 = Completed in default CiviCRM
  })
  is_active = true
  negate    = false
}

# Action: change case status to "Housing Secured"
# action_id: ID of "Change case status" action from CiviRulesAction.get
resource "civicrm_civirules_rule_action" "set_status_housing_secured" {
  rule_id   = civicrm_civirules_rule.income_stab_closes_case.id
  action_id = 8
  action_params = jsonencode({
    case_status_id = civicrm_case_status.housing_secured.value
  })
  is_active = true
}

# ─────────────────────────────────────────────────────────────────────────────
# Second rule: close case when final follow-up is completed
# ─────────────────────────────────────────────────────────────────────────────

resource "civicrm_civirules_rule" "followup_closes_case" {
  name       = "housing_final_followup_close"
  label      = "Housing: Final Follow-up completed → Close Case"
  trigger_id = 5
  description = "When the final follow-up activity is completed, change the case status to Case Closed."
  is_active  = true
}

resource "civicrm_civirules_rule_condition" "check_followup_type" {
  rule_id      = civicrm_civirules_rule.followup_closes_case.id
  condition_id = 3
  condition_params = jsonencode({
    activity_type_id = "Follow up"
  })
  is_active = true
}

resource "civicrm_civirules_rule_condition" "check_followup_status" {
  rule_id      = civicrm_civirules_rule.followup_closes_case.id
  condition_id = 4
  condition_params = jsonencode({
    activity_status_id = "2"
  })
  is_active = true
}

resource "civicrm_civirules_rule_action" "close_case" {
  rule_id   = civicrm_civirules_rule.followup_closes_case.id
  action_id = 8
  action_params = jsonencode({
    case_status_id = civicrm_case_status.case_closed.value
  })
  is_active = true
}
