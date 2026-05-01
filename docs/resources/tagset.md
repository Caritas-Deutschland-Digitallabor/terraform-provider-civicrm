---
page_title: "civicrm_tagset Resource - CiviCRM"
subcategory: ""
description: |-
  Manages a CiviCRM Tagset. A Tagset is a named container for tags.
---

# civicrm_tagset (Resource)

Manages a CiviCRM Tagset. A Tagset is a named container (group) for tags. Individual tags reference their tagset via `parent_id` in the `civicrm_tag` resource.

Internally, a Tagset is a Tag entity with `is_tagset = true`. This resource sets that flag automatically.

## Example Usage

```terraform
resource "civicrm_tagset" "contact_categories" {
  name     = "Contact_Categories"
  label    = "Contact Categories"
  used_for = ["civicrm_contact"]
}

# Tags belonging to this tagset
resource "civicrm_tag" "vip" {
  name      = "VIP"
  label     = "VIP"
  parent_id = civicrm_tagset.contact_categories.id
  used_for  = ["civicrm_contact"]
  color     = "#ffd700"
}
```

## Argument Reference

### Required

- `name` (String) The machine name of the tagset (must be unique, no spaces).

### Optional

- `label` (String) The display label of the tagset. Defaults to `name` if not set.
- `description` (String) A description of the tagset.
- `used_for` (List of String) Entity types this tagset can be used for (e.g., `civicrm_contact`, `civicrm_activity`).
- `color` (String) The color for the tagset in hex format (e.g., `#ff0000`).

## Attributes Reference

- `id` (Number) The unique identifier of the tagset.

## Import

Tagsets can be imported using the tag ID:

```shell
terraform import civicrm_tagset.example 42
```
