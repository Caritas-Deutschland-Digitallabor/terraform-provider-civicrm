# Case roles are standard CiviCRM RelationshipTypes.
# There is no separate civicrm_case_role resource — use civicrm_relationship_type.
#
# CiviCase references roles by name_a_b in the case type definition.
# The role contact type restrictions should be left empty (null) so the role
# works for all contact types, which is the CiviCase convention.

resource "civicrm_relationship_type" "housing_coordinator" {
  name_a_b = "Housing Coordinator"
  label_a_b = "Housing Coordinator"
  name_b_a = "Housing Coordinator of"
  label_b_a = "Housing Coordinator of"
  is_active = true
}

resource "civicrm_relationship_type" "social_worker" {
  name_a_b = "Social Worker"
  label_a_b = "Sozialarbeiter/in"
  name_b_a = "Social Worker of"
  label_b_a = "Betreute Person von Sozialarbeiter/in"
  is_active = true
}

# Reference from a case type — use name_a_b as the role name:
#
# resource "civicrm_case_type" "housing_support" {
#   name  = "housing_support"
#   title = "Wohnungsunterstützung"
#
#   definition = jsonencode({
#     activityTypes = [...]
#     activitySets  = [...]
#     caseRoles = [
#       { name = civicrm_relationship_type.housing_coordinator.name_a_b, creator = "1", manager = "1" },
#       { name = civicrm_relationship_type.social_worker.name_a_b },
#     ]
#   })
# }
