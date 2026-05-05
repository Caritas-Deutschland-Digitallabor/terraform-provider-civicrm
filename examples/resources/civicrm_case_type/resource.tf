# Minimal case type without a definition
resource "civicrm_case_type" "simple" {
  name        = "my_case_type"
  title       = "My Case Type"
  description = "A simple case type"
  is_active   = true
}

# Case type with a full definition (activity types, timeline, roles)
resource "civicrm_case_type" "housing_support" {
  name        = "housing_support"
  title       = "Housing Support"
  description = "Supports homeless people in finding housing"
  is_active   = true
  is_reserved = false
  weight      = 1

  definition = jsonencode({
    activityTypes = [
      { name = "Open Case", max_instances = "1" },
      { name = "Follow up" },
      { name = "Change Case Status" }
    ]
    activitySets = [
      {
        name     = "standard_timeline"
        label    = "Standard Timeline"
        timeline = 1
        activityTypes = [
          { name = "Open Case", status = "Completed" },
          {
            name               = "Follow up"
            reference_activity = "Open Case"
            reference_offset   = "7"
            reference_select   = "newest"
          }
        ]
      }
    ]
    caseRoles = [
      { name = "Housing Coordinator", creator = "1", manager = "1" }
    ]
  })
}
