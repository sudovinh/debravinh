# State lives in the shared Spaces bucket (see terraform-root/digitalocean/README.md).
# Credentials come from AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY via `op run` (Makefile).
terraform {
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
}
