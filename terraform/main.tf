terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

resource "digitalocean_app" "debravinh_web" {
  spec {
    name   = "debravinh-com"
    region = "sfo"

    alert {
      rule = "DEPLOYMENT_FAILED"
    }

    alert {
      rule = "DOMAIN_FAILED"
    }

    domain {
      name = "debravinh.com"
      type = "PRIMARY"
      zone = "debravinh.com"
    }

    domain {
      name = "www.debravinh.com"
      type = "ALIAS"
      zone = "debravinh.com"
    }

    service {
      name               = "sudovinh-debravinh"
      instance_count     = 1
      instance_size_slug = "basic-xxs"

      github {
        repo           = "sudovinh/debravinh"
        branch         = "main"
        deploy_on_push = false
      }

      dockerfile_path = "Dockerfile"

      http_port = 8080

      health_check {
        http_path             = "/"
        initial_delay_seconds = 10
        period_seconds        = 30
      }
    }
  }
}

# DNS zone for debravinh.com lives in DigitalOcean and points at the app.
resource "digitalocean_domain" "debravinh" {
  name = "debravinh.com"
}

# Existing resources are imported into state, not recreated.
import {
  to = digitalocean_app.debravinh_web
  id = "90e1359d-4e42-47a0-817c-63a83df75eb5"
}

import {
  to = digitalocean_domain.debravinh
  id = "debravinh.com"
}

output "app_url" {
  value       = digitalocean_app.debravinh_web.live_url
  description = "The live URL of the deployed app"
}

output "default_ingress" {
  value       = digitalocean_app.debravinh_web.default_ingress
  description = "The default ondigitalocean.app ingress for the app"
}
