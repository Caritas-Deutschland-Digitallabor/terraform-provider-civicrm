resource "civicrm_civirules_trigger" "weekly_check" {
  name       = "my_weekly_housing_check"
  label      = "Weekly Housing Status Check"
  cron       = true
  class_name = "CRM_MyExtension_Trigger_WeeklyCheck"
  is_active  = true
}
