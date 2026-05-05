# Look up an existing relationship type by name to use as a case role
data "civicrm_relationship_type" "housing_coordinator" {
  name_a_b = "Housing Coordinator"
}

# Reference in a case type definition
resource "civicrm_case_type" "housing_support" {
  name  = "housing_support"
  title = "Wohnungsunterstützung"

  definition = jsonencode({
    activityTypes = [{ name = "Open Case", max_instances = "1" }]
    activitySets  = []
    caseRoles = [
      {
        name    = data.civicrm_relationship_type.housing_coordinator.name_a_b
        creator = "1"
        manager = "1"
      }
    ]
  })
}
