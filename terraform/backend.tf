# State lives in the shared Spaces bucket (see terraform-root/digitalocean/README.md),
# encrypted client-side (OpenTofu only — run `tofu`, not `terraform`). Spaces creds +
# passphrase come from .env.terraform via `op run` (Makefile).
variable "state_passphrase" {
  description = "OpenTofu state/plan encryption passphrase (TF_VAR_state_passphrase via op run)"
  type        = string
  sensitive   = true
}

terraform {
  required_version = ">= 1.8"

  backend "s3" {
    endpoints = {
      s3 = "https://sfo3.digitaloceanspaces.com"
    }
    bucket       = "sv-terraform-root"
    key          = "apps/debravinh/terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true

    skip_credentials_validation = true
    skip_requesting_account_id  = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
    skip_s3_checksum            = true
  }

  encryption {
    key_provider "pbkdf2" "state" {
      passphrase = var.state_passphrase
    }
    method "aes_gcm" "state" {
      keys = key_provider.pbkdf2.state
    }
    # MIGRATION PHASE: reads the existing plaintext state once. OpenTofu
    # forbids `enforced = true` alongside an unencrypted fallback, so the
    # follow-up drops this method + fallback and sets enforced on state.
    method "unencrypted" "migrate" {}

    state {
      method = method.aes_gcm.state
      fallback {
        method = method.unencrypted.migrate
      }
    }
    plan {
      method   = method.aes_gcm.state
      enforced = true
    }
  }
}
