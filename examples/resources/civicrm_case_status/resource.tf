# Open/active statuses use grouping = "Opened"
# Resolved/ended statuses use grouping = "Closed"

resource "civicrm_case_status" "initial_assessment" {
  name     = "initial_assessment"
  label    = "Initial Assessment"
  grouping = "Opened"
  weight   = 10
}

resource "civicrm_case_status" "in_progress" {
  name     = "in_progress"
  label    = "In Progress"
  grouping = "Opened"
  weight   = 20
}

resource "civicrm_case_status" "housing_secured" {
  name     = "housing_secured"
  label    = "Housing Secured"
  grouping = "Closed"
  weight   = 30
}

resource "civicrm_case_status" "case_closed" {
  name     = "case_closed"
  label    = "Case Closed"
  grouping = "Closed"
  weight   = 40
}
