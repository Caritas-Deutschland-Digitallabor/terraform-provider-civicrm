# A tagset for categorizing contacts
resource "civicrm_tagset" "contact_categories" {
  name     = "Contact_Categories"
  label    = "Contact Categories"
  used_for = ["civicrm_contact"]
}

# Tags that belong to the tagset reference it via parent_id in civicrm_tag
resource "civicrm_tag" "vip" {
  name      = "VIP"
  label     = "VIP"
  parent_id = civicrm_tagset.contact_categories.id
  used_for  = ["civicrm_contact"]
  color     = "#gold"
}
