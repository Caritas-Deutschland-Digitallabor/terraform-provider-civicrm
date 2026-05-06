# Terraform Provider for CiviCRM

[![Build](https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/actions/workflows/build.yml/badge.svg)](https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/actions/workflows/build.yml)
[![Release](https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/actions/workflows/release.yml/badge.svg)](https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/actions/workflows/release.yml)

A Terraform provider for managing CiviCRM access control resources via API v4.

## Quick Start

```hcl
terraform {
  required_providers {
    civicrm = {
      source  = "Caritas-Deutschland-Digitallabor/civicrm"
      version = "~> 0.1"
    }
  }
}

provider "civicrm" {
  url     = "https://your-civicrm-instance.org"
  api_key = "your-api-key"
}

resource "civicrm_group" "volunteers" {
  name  = "volunteers"
  title = "Volunteers"
}
```

Run `terraform init` to download the provider, then `terraform apply` to create resources.

> **Note**: This provider is distributed via GitHub releases, not the Terraform Registry.  
> See [GITHUB-DISTRIBUTION.md](GITHUB-DISTRIBUTION.md) for details on how this works.

## Features

This provider supports managing the following CiviCRM resources:

- **Groups** (`civicrm_group`) - CiviCRM groups that can be assigned ACL roles
- **ACL Roles** (`civicrm_acl_role`) - Permission roles that define access levels
- **ACLs** (`civicrm_acl`) - Access control rules defining what operations a role can perform
- **ACL Entity Roles** (`civicrm_acl_entity_role`) - Assigns ACL roles to groups (role bindings)
- **Custom Groups** (`civicrm_custom_group`) - Custom field sets attached to CiviCRM entities
- **Custom Fields** (`civicrm_custom_field`) - Individual custom fields within a custom group
- **Tags** (`civicrm_tag`) - Tags for categorising contacts and other entities
- **Tagsets** (`civicrm_tagset`) - Named collections of tags
- **Contact Types** (`civicrm_contact_type`) - Custom contact subtypes
- **Relationship Types** (`civicrm_relationship_type`) - Named relationship types between contacts
- **Case Types** (`civicrm_case_type`) - Case structures with activity types and case roles
- **Case Statuses** (`civicrm_case_status`) - Status options for cases
- **Membership Types** (`civicrm_membership_type`) - Membership definitions including fees, duration, and renewal rules
- **Mail Settings** (`civicrm_mail_settings`) - Inbound mail account configuration
- **Site Email Addresses** (`civicrm_site_email_address`) - From-addresses used by CiviCRM when sending mail
- **CiviRules Rules** (`civicrm_civirules_rule`) - Automation rules triggered by CiviCRM events
- **CiviRules Triggers** (`civicrm_civirules_trigger`) - Trigger definitions for CiviRules
- **CiviRules Rule Conditions** (`civicrm_civirules_rule_condition`) - Conditions attached to CiviRules rules
- **CiviRules Rule Actions** (`civicrm_civirules_rule_action`) - Actions executed by CiviRules rules

Each resource also has a corresponding data source for read-only lookups.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (only for building from source)
- CiviCRM >= 5.47 (required for full REST API v4 support)

## Using the Provider

### From GitHub Releases (Without Terraform Registry)

This provider is distributed via **GitHub releases** and does not require registration with the official Terraform Registry. Terraform can automatically download providers from GitHub using the implicit provider installation mechanism.

#### How It Works

When you specify a provider source like `Caritas-Deutschland-Digitallabor/civicrm`, Terraform interprets this as:

1. **Hostname**: `registry.terraform.io` (default when not specified)
2. **Namespace**: `Caritas-Deutschland-Digitallabor`
3. **Type**: `civicrm`

However, since this provider is not published to the official Terraform Registry, Terraform falls back to looking for releases directly from the GitHub repository at:
```
https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/releases
```

Terraform automatically:
- Detects the GitHub repository from the namespace
- Downloads the appropriate binary for your platform from GitHub releases
- Verifies checksums (SHA256SUMS file)
- Verifies GPG signatures (if available)
- Caches the provider locally

#### Basic Usage

To use this provider in your Terraform configuration, specify it in your `required_providers` block:

```hcl
terraform {
  required_providers {
    civicrm = {
      source  = "Caritas-Deutschland-Digitallabor/civicrm"
      version = "~> 0.1"
    }
  }
}

provider "civicrm" {
  url     = "https://your-civicrm-instance.org"
  api_key = "your-api-key"
}
```

Terraform will automatically download the provider from the GitHub releases when you run `terraform init`.

### Provider Installation

When using the provider from GitHub releases, Terraform handles the installation automatically. The provider supports the following platforms:

- Linux (amd64, arm64)
- macOS/Darwin (amd64, arm64)  
- Windows (amd64, arm64)

### Using in Your Terraform Projects

To use this provider in your Terraform projects:

1. **Create a new Terraform configuration** or add to an existing one:

```hcl
# main.tf
terraform {
  required_version = ">= 1.0"
  
  required_providers {
    civicrm = {
      source  = "Caritas-Deutschland-Digitallabor/civicrm"
      version = "~> 0.1"
    }
  }
}

provider "civicrm" {
  url     = var.civicrm_url
  api_key = var.civicrm_api_key
}

# Your CiviCRM resources here
resource "civicrm_group" "example" {
  name  = "example_group"
  title = "Example Group"
}
```

2. **Set up variables** (optional but recommended):

```hcl
# variables.tf
variable "civicrm_url" {
  description = "The URL of your CiviCRM instance"
  type        = string
}

variable "civicrm_api_key" {
  description = "The API key for CiviCRM authentication"
  type        = string
  sensitive   = true
}
```

3. **Initialize and apply**:

```bash
# Initialize Terraform (downloads the provider)
terraform init

# Plan your changes
terraform plan

# Apply the configuration
terraform apply
```

4. **Pass credentials securely**:

```bash
# Option 1: Use environment variables
export TF_VAR_civicrm_url="https://your-instance.org"
export TF_VAR_civicrm_api_key="your-api-key"
terraform apply

# Option 2: Use a terraform.tfvars file (add to .gitignore!)
cat > terraform.tfvars <<EOF
civicrm_url     = "https://your-instance.org"
civicrm_api_key = "your-api-key"
EOF
terraform apply

# Option 3: Use command-line flags
terraform apply \
  -var="civicrm_url=https://your-instance.org" \
  -var="civicrm_api_key=your-api-key"
```

**Important Security Notes:**
- Never commit `terraform.tfvars` files containing secrets
- Use environment variables or secret management tools in CI/CD
- Consider using the provider's environment variable support: `CIVICRM_URL` and `CIVICRM_API_KEY`


## Building the Provider from Source

If you need to build the provider from source (for development or custom modifications):

1. Clone the repository:
```bash
git clone https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform.git
cd civicrm-terraform
```

2. Build the provider:
```bash
go mod tidy
go build -o terraform-provider-civicrm
```

3. Install for local development:
```bash
# Linux/Mac
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/Caritas-Deutschland-Digitallabor/civicrm/0.1.0/$(go env GOOS)_$(go env GOARCH)
cp terraform-provider-civicrm ~/.terraform.d/plugins/registry.terraform.io/Caritas-Deutschland-Digitallabor/civicrm/0.1.0/$(go env GOOS)_$(go env GOARCH)/

# Windows (PowerShell)
$installDir = "$env:APPDATA\terraform.d\plugins\registry.terraform.io\Caritas-Deutschland-Digitallabor\civicrm\0.1.0\windows_amd64"
New-Item -ItemType Directory -Force -Path $installDir
Copy-Item terraform-provider-civicrm.exe $installDir
```

## GitHub-Based Distribution (Not Using Official Terraform Registry)

This provider is distributed through **GitHub Releases** and is designed to work without being published to the official Terraform Registry. This section explains how this works and what you need to know.

### Understanding the Distribution Model

When you reference a provider with the source `Caritas-Deutschland-Digitallabor/civicrm`, Terraform uses its **implicit provider installation** mechanism:

1. **Provider Source Format**: `<NAMESPACE>/<TYPE>`
   - In this case: `Caritas-Deutschland-Digitallabor/civicrm`
   - No hostname means it defaults to `registry.terraform.io`

2. **GitHub Repository Discovery**:
   - Terraform recognizes organization names from GitHub
   - It looks for releases at: `https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/releases`
   - The repository name typically follows the pattern `terraform-provider-<TYPE>` but Terraform can also find repos named differently

3. **Release Requirements**:
   - Releases must be tagged with semantic versioning (e.g., `v0.1.0`)
   - Must include platform-specific binaries (e.g., `terraform-provider-civicrm_v0.1.0_linux_amd64.zip`)
   - Must include a `<PROJECT>_<VERSION>_SHA256SUMS` file with checksums
   - Optionally includes a `<PROJECT>_<VERSION>_SHA256SUMS.sig` GPG signature file

### What Happens During `terraform init`

When you run `terraform init`, Terraform:

1. **Parses** your `required_providers` block
2. **Queries** for available versions (from GitHub API if not in registry)
3. **Selects** a version that matches your constraint (e.g., `~> 0.1`)
4. **Downloads** the appropriate binary for your platform:
   - Detects your OS and architecture
   - Downloads the ZIP archive from GitHub releases
   - Example: `terraform-provider-civicrm_v0.1.0_linux_amd64.zip`
5. **Verifies** the download:
   - Checks SHA256 checksum against the `SHA256SUMS` file
   - Verifies GPG signature if present and configured
6. **Installs** the provider to your local plugin cache:
   - Linux/Mac: `~/.terraform.d/plugins/`
   - Windows: `%APPDATA%\terraform.d\plugins\`
7. **Caches** the provider for future use

### Advantages of GitHub-Based Distribution

✅ **No Registry Registration Required**: No need to publish to Terraform Registry  
✅ **Automatic Distribution**: Works with existing CI/CD (GitHub Actions + GoReleaser)  
✅ **Version Control**: Versions tied directly to Git tags  
✅ **Simple Publishing**: Just push a version tag to trigger a release  
✅ **Open Source Friendly**: Perfect for open-source projects  
✅ **Standard Terraform Workflow**: Users use standard `terraform init`

### Comparison: GitHub vs Official Registry

| Aspect | GitHub Releases | Official Terraform Registry |
|--------|----------------|----------------------------|
| **Setup** | Push version tags | Sign up, verify namespace, configure |
| **Publishing** | Automatic via GitHub Actions | Manual or CI/CD to registry |
| **Discovery** | Direct GitHub organization reference | Searchable on registry.terraform.io |
| **Documentation** | GitHub README/Wiki | Registry documentation page |
| **Provider Source** | `org/provider` | `org/provider` (same!) |
| **Usage** | Identical for end users | Identical for end users |

### Requirements for GitHub-Based Providers

To successfully distribute a Terraform provider via GitHub releases, ensure:

1. **Repository Naming**: 
   - Recommended: `terraform-provider-<name>` (e.g., `terraform-provider-civicrm`)
   - Alternative: Repository name should contain the provider type

2. **Release Artifacts**:
   - Binary naming: `terraform-provider-<name>_v<version>`
   - Archive naming: `terraform-provider-<name>_<version>_<os>_<arch>.zip`
   - Checksum file: `terraform-provider-<name>_<version>_SHA256SUMS`
   - Optional signature: `terraform-provider-<name>_<version>_SHA256SUMS.sig`

3. **Version Tags**:
   - Must use semantic versioning: `v0.1.0`, `v1.2.3`, etc.
   - Tags trigger automated releases (via GoReleaser)

4. **Platform Binaries**:
   - Must support common platforms: linux_amd64, darwin_amd64, windows_amd64
   - Optional: Additional platforms (arm64, etc.)

### Troubleshooting GitHub-Based Installation

#### Provider Not Found

If Terraform can't find the provider:

```
Error: Failed to query available provider packages
│ Could not retrieve the list of available versions for provider Caritas-Deutschland-Digitallabor/civicrm
```

**Solutions**:
1. Verify a release exists: https://github.com/Caritas-Deutschland-Digitallabor/civicrm-terraform/releases
2. Check that the release is not a draft
3. Ensure the release is tagged with `v` prefix (e.g., `v0.1.0`)
4. Verify you have internet access to GitHub

#### Checksum Verification Failed

If checksum verification fails:

```
Error: Failed to install provider
│ Error while installing Caritas-Deutschland-Digitallabor/civicrm: checksum mismatch
```

**Solutions**:
1. Clear Terraform's plugin cache: `rm -rf ~/.terraform.d/plugins/`
2. Run `terraform init` again
3. Check if the release was re-uploaded (GitHub releases are immutable, but can be deleted/recreated)

#### Wrong Platform Binary

If Terraform downloads the wrong binary:

**Solutions**:
1. Check available platforms in the release
2. Ensure your platform is supported
3. For uncommon platforms, you may need to build from source

### Using in Air-Gapped Environments

For environments without internet access, you can manually install the provider:

1. **Download** the appropriate binary for your platform from GitHub releases
2. **Extract** the ZIP archive
3. **Place** the binary in the plugin directory:
   ```bash
   # Linux/macOS
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/Caritas-Deutschland-Digitallabor/civicrm/<VERSION>/<OS>_<ARCH>/
   cp terraform-provider-civicrm_v<VERSION> ~/.terraform.d/plugins/registry.terraform.io/Caritas-Deutschland-Digitallabor/civicrm/<VERSION>/<OS>_<ARCH>/
   
   # Example for Linux AMD64, version 0.1.0:
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/Caritas-Deutschland-Digitallabor/civicrm/0.1.0/linux_amd64/
   cp terraform-provider-civicrm_v0.1.0 ~/.terraform.d/plugins/registry.terraform.io/Caritas-Deutschland-Digitallabor/civicrm/0.1.0/linux_amd64/
   ```

4. **Run** `terraform init` - it will use the local binary

Alternatively, use a [Terraform provider network mirror](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation) for managing multiple air-gapped installations.

## Configuration

### Provider Configuration

```hcl
provider "civicrm" {
  url     = "https://your-civicrm-instance.org"  # Required
  api_key = "your-api-key"                        # Required

  # Optional
  insecure = false  # Skip TLS verification (for development only)
}
```

### Environment Variables

The provider supports the following environment variables:

| Variable | Description |
|----------|-------------|
| `CIVICRM_URL` | The base URL of your CiviCRM instance |
| `CIVICRM_API_KEY` | Your CiviCRM API key |

### CiviCRM Setup

1. Ensure your CiviCRM instance is version 5.47 or later.

2. Enable the AuthX extension if not already enabled.

3. Generate an API key for a user with appropriate permissions:
   - Navigate to **Administer > Users and Permissions > API Key**
   - Or use cv: `cv api4 Contact.create '{"id": YOUR_CONTACT_ID, "api_key": "your-generated-key"}'`

4. Ensure the user has the following permissions:
   - "access CiviCRM"
   - "authenticate with api key"
   - Appropriate ACL permissions for managing groups and ACLs

## Usage Examples

### Create Groups and ACL Roles

```hcl
# Create a group for team members
resource "civicrm_group" "team_leaders" {
  name        = "team_leaders"
  title       = "Team Leaders"
  description = "Staff members with team lead responsibilities"
  is_active   = true
}

# Create a group for volunteers
resource "civicrm_group" "volunteers" {
  name        = "volunteers"
  title       = "Volunteers"
  description = "Active volunteers"
  is_active   = true
}

# Create an ACL role
resource "civicrm_acl_role" "volunteer_manager" {
  name        = "volunteer_manager"
  label       = "Volunteer Manager"
  description = "Can view and edit volunteers"
  is_active   = true
}
```

### Create ACL Rules

```hcl
# Grant Edit permission on volunteers group to volunteer_manager role
resource "civicrm_acl" "vm_edit_volunteers" {
  name         = "vm_edit_volunteers"
  entity_id    = civicrm_acl_role.volunteer_manager.id
  operation    = "Edit"
  object_table = "civicrm_group"
  object_id    = civicrm_group.volunteers.id
  is_active    = true
}
```

### Assign Roles to Groups

```hcl
# Give team leaders the volunteer manager role
resource "civicrm_acl_entity_role" "team_leaders_vm" {
  acl_role_id  = civicrm_acl_role.volunteer_manager.id
  entity_table = "civicrm_group"
  entity_id    = civicrm_group.team_leaders.id
  is_active    = true
}
```

### Look Up Existing Resources

```hcl
# Find an existing group by name
data "civicrm_group" "administrators" {
  name = "Administrators"
}

# Find an existing ACL role by name
data "civicrm_acl_role" "admin_role" {
  name = "Administrator"
}
```

## Resource Reference

### civicrm_group

Manages a CiviCRM Group.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | The machine name of the group (must be unique) |
| `title` | string | Yes | The display title of the group |
| `description` | string | No | A description of the group |
| `is_active` | bool | No | Whether the group is active (default: true) |
| `visibility` | string | No | Visibility setting (default: "User and User Admin Only") |
| `group_type` | list(string) | No | The types of the group |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | int | The unique identifier of the group |

### civicrm_acl_role

Manages a CiviCRM ACL Role.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | The machine name of the ACL role |
| `label` | string | Yes | The display label of the ACL role |
| `description` | string | No | A description of the ACL role |
| `is_active` | bool | No | Whether the ACL role is active (default: true) |
| `weight` | int | No | The sort weight of the ACL role |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | int | The unique identifier of the ACL role |
| `value` | string | The internal value (auto-generated) |

### civicrm_acl

Manages a CiviCRM ACL rule.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | The name of the ACL rule |
| `entity_id` | int | Yes | The ID of the ACL role this rule belongs to |
| `operation` | string | Yes | The operation: "Edit", "View", "Create", "Delete", "Search", "All" |
| `object_table` | string | Yes | The type of object being permissioned |
| `entity_table` | string | No | The entity table (default: "civicrm_acl_role") |
| `object_id` | int | No | The specific object ID (null = all) |
| `is_active` | bool | No | Whether the ACL rule is active (default: true) |
| `deny` | bool | No | Deny instead of allow (default: false) |
| `priority` | int | No | Rule priority |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | int | The unique identifier of the ACL |

### civicrm_acl_entity_role

Manages a CiviCRM ACL Entity Role assignment.

#### Arguments

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `acl_role_id` | int | Yes | The ID of the ACL role to assign |
| `entity_id` | int | Yes | The ID of the group to assign the role to |
| `entity_table` | string | No | The entity table (default: "civicrm_group") |
| `is_active` | bool | No | Whether the assignment is active (default: true) |

#### Attributes

| Name | Type | Description |
|------|------|-------------|
| `id` | int | The unique identifier of the assignment |

## Importing Existing Resources

All resources support importing by ID:

```bash
# Import a group
terraform import civicrm_group.example 123

# Import an ACL role
terraform import civicrm_acl_role.example 456

# Import an ACL
terraform import civicrm_acl.example 789

# Import an ACL entity role
terraform import civicrm_acl_entity_role.example 101
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Local Installation

```bash
make install
```

### Code Formatting

```bash
make fmt
```

## Additional Documentation

- **[GITHUB-DISTRIBUTION.md](GITHUB-DISTRIBUTION.md)** - Understanding GitHub-based provider distribution (not using Terraform Registry)
- **[USAGE.md](USAGE.md)** - Comprehensive guide for using this provider in your projects
- **[PUBLISHING.md](PUBLISHING.md)** - Guide for maintainers on publishing releases
- **[CHANGELOG.md](CHANGELOG.md)** - Version history and release notes

## License

This project is licensed under the MIT License - see the LICENSE file for details.
