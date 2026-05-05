---
page_title: "civicrm_case_type Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM Case Type. Case Types define the structure of cases including activity types, timelines, and case roles.
---

# civicrm_case_type (Resource)

Manages a CiviCRM Case Type. Case Types define the structure of cases, including which activity types are available, timeline schedules, and the roles that participants play in a case.

## Example Usage

```terraform
# Minimal case type without a definition
resource "civicrm_case_type" "simple" {
  name        = "my_case_type"
  title       = "My Case Type"
  description = "A simple case type"
  is_active   = true
}

# Case type with a full definition
resource "civicrm_case_type" "housing_support" {
  name        = "housing_support"
  title       = "Housing Support"
  description = "Supports homeless people in finding housing"
  is_active   = true
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
```

## Argument Reference

The following arguments are supported:

### Required

- `name` (String) The machine name of the case type (must be unique).
- `title` (String) The display title of the case type.

### Optional

- `description` (String) A description of the case type.
- `is_active` (Boolean) Whether the case type is active. Default: `true`.
- `is_reserved` (Boolean) Whether the case type is reserved (system-defined). Default: `false`.
- `weight` (Number) The sort weight of the case type.
- `definition` (String) The case type definition as a JSON string. Use `jsonencode()` to construct it. The definition can contain:
  - `activityTypes` — list of activity types available in this case type
  - `activitySets` — named sets of activities, including timeline schedules
  - `timelineActivityTypes` — flat list of timeline activities (computed by CiviCRM from activitySets)
  - `caseRoles` — roles that participants can hold in this case

  On read, the API returns a structured object which is serialized back to a JSON string in state.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` (Number) The unique identifier of the case type.
- `definition` (String) The case type definition as a JSON string (computed if not specified).
- `weight` (Number) The sort weight (computed if not specified).

## Import

Case Types can be imported using the case type ID:

```shell
terraform import civicrm_case_type.example 1
```
