/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

terraform {
  backend "gcs" {
    bucket = "pi-delivery-tf-state"
    prefix = "terraform/state"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.47.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 4.47.0"
    }
  }
}

provider "google" {
  project = var.project
}

provider "google-beta" {
  project = var.project
}

locals {
  managed_domains = [
    "api.staging.pi.delivery"
  ]
  regions = toset(var.regions)
}

data "google_cloudfunctions2_function" "api_pi" {
  for_each = local.regions
  name     = "api-pi"
  location = each.value
}

data "google_cloudfunctions2_function" "api_not_found" {
  name     = "api-not-found"
  location = "us-central1"
}

data "google_cloudfunctions2_function" "api_pi_staging" {
  name     = "api-pi-staging"
  location = "us-central1"
}

resource "google_storage_bucket" "functions_staging" {
  name     = "piaas-gcp-gcf-staging"
  location = "US"

  uniform_bucket_level_access = true
}

resource "google_service_account" "functions_api" {
  account_id   = "sa-functions-api"
  display_name = "Service Account for API on Cloud Functions"
}

resource "google_project_iam_binding" "storage_object_viewer" {
  project = var.project
  role    = "roles/storage.objectViewer"

  members = [
    "serviceAccount:${google_service_account.functions_api.email}"
  ]
}

resource "google_project_iam_binding" "logging_log_writer" {
  project = var.project
  role    = "roles/logging.logWriter"

  members = [
    "serviceAccount:${google_service_account.functions_api.email}"
  ]
}

resource "google_project_iam_binding" "monitoring_metric_writer" {
  project = var.project
  role    = "roles/monitoring.metricWriter"

  members = [
    "serviceAccount:${google_service_account.functions_api.email}"
  ]
}

resource "google_compute_global_address" "api" {
  name = "global-api-ip"
}

resource "google_compute_global_address" "api_v6" {
  name       = "global-api-ip-v6"
  ip_version = "IPV6"
}

resource "random_id" "certificate" {
  byte_length = 4
  prefix      = "cert-"
  keepers = {
    managed_domains = join(",", local.managed_domains)
  }
}

resource "google_compute_managed_ssl_certificate" "certificate" {
  name = random_id.certificate.hex
  managed {
    domains = local.managed_domains
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "random_id" "neg" {
  byte_length = 4
  keepers = {
    v = 7
  }
}

resource "google_compute_region_network_endpoint_group" "api_func_pi_prod" {
  for_each = data.google_cloudfunctions2_function.api_pi

  name                  = "api-neg-func-pi-prod-${each.value.location}-${random_id.neg.hex}"
  network_endpoint_type = "SERVERLESS"
  region                = each.value.location

  description = "API network endpoint for /v1/pi ${each.value.location}"

  cloud_run {
    service = each.value.name
  }
}

resource "google_compute_region_network_endpoint_group" "api_func_pi_staging" {
  name                  = "api-neg-func-pi-staging-${random_id.neg.hex}"
  network_endpoint_type = "SERVERLESS"
  region                = data.google_cloudfunctions2_function.api_pi_staging.location

  description = "API network endpoint for /v1/pi staging"

  cloud_run {
    service = data.google_cloudfunctions2_function.api_pi_staging.name
  }
}

resource "google_compute_region_network_endpoint_group" "api_not_found" {
  name                  = "api-neg-not-found"
  network_endpoint_type = "SERVERLESS"
  region                = data.google_cloudfunctions2_function.api_not_found.location
  description           = "Endpoint for 404"

  cloud_run {
    service = data.google_cloudfunctions2_function.api_not_found.name
  }
}

resource "random_id" "backend" {
  byte_length = 4
}

resource "google_compute_backend_service" "api_func_pi_prod" {
  name                  = "api-func-pi-prod-backend-${random_id.backend.hex}"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  security_policy       = "api-prod"

  dynamic "backend" {
    for_each = google_compute_region_network_endpoint_group.api_func_pi_prod
    content {
      group = backend.value.id
    }
  }

  log_config {
    enable      = true
    sample_rate = 0.2
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_backend_service" "api_func_pi_staging" {
  name                  = "api-func-pi-staging-backend-${random_id.backend.hex}"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  security_policy       = "api-staging"

  backend {
    group = google_compute_region_network_endpoint_group.api_func_pi_staging.id
  }

  log_config {
    enable      = true
    sample_rate = 1.0
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_backend_service" "api_not_found" {
  name                  = "api-not-found-backend"
  load_balancing_scheme = "EXTERNAL_MANAGED"
  security_policy       = "api-prod"

  backend {
    group = google_compute_region_network_endpoint_group.api_not_found.id
  }

  log_config {
    enable      = true
    sample_rate = 1.0
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "google_compute_url_map" "api" {
  name = "api-url-map"

  default_service = google_compute_backend_service.api_not_found.id
  default_route_action {
    cors_policy {
      allow_origins = ["*"]
      disabled      = false
    }
  }

  header_action {
    response_headers_to_add {
      header_name  = "Strict-Transport-Security"
      header_value = "max-age=63072000"
      replace      = true
    }

    response_headers_to_add {
      header_name  = "Access-Control-Allow-Origin"
      header_value = "*"
      replace      = true
    }
  }

  host_rule {
    hosts        = ["api.pi.delivery"]
    path_matcher = "api-func-prod"
  }

  host_rule {
    hosts        = ["api.staging.pi.delivery"]
    path_matcher = "api-func-staging"
  }

  path_matcher {
    name            = "api-func-prod"
    default_service = google_compute_backend_service.api_not_found.id
    path_rule {
      paths   = ["/v1/pi"]
      service = google_compute_backend_service.api_func_pi_prod.id
      route_action {
        cors_policy {
          allow_origins = ["*"]
          disabled      = false
        }
      }
    }

    default_route_action {
      cors_policy {
        allow_origins = ["*"]
        disabled      = false
      }
    }
  }

  path_matcher {
    name            = "api-func-staging"
    default_service = google_compute_backend_service.api_not_found.id
    path_rule {
      paths   = ["/v1/pi"]
      service = google_compute_backend_service.api_func_pi_staging.id
      route_action {
        cors_policy {
          allow_origins = ["*"]
          disabled      = false
        }
      }
    }

    default_route_action {
      cors_policy {
        allow_origins = ["*"]
        disabled      = false
      }
    }
  }
}

resource "google_compute_managed_ssl_certificate" "api" {
  name = "cert-api-pi-delivery"
  managed {
    domains = ["api.staging.pi.delivery", "api.pi.delivery"]
  }
  lifecycle {
    prevent_destroy = true
  }
}

data "google_compute_ssl_certificate" "k8s" {
  name = "mcrt-4dd17c3e-570e-4db8-9ec5-45803b394fd0"
}

resource "google_compute_ssl_policy" "ssl_policy" {
  name            = "ssl-policy"
  profile         = "MODERN"
  min_tls_version = "TLS_1_2"
}

resource "google_compute_target_https_proxy" "api" {
  name    = "api-https-proxy"
  url_map = google_compute_url_map.api.id
  ssl_certificates = [google_compute_managed_ssl_certificate.certificate.id,
    google_compute_managed_ssl_certificate.api.id,
  data.google_compute_ssl_certificate.k8s.id]
  ssl_policy = google_compute_ssl_policy.ssl_policy.id
}

resource "google_compute_global_forwarding_rule" "api" {
  name       = "api-forwarding-rule"
  port_range = "443"

  load_balancing_scheme = "EXTERNAL_MANAGED"
  ip_address            = google_compute_global_address.api.id
  target                = google_compute_target_https_proxy.api.id
}

resource "google_compute_global_forwarding_rule" "api_v6" {
  name       = "api-forwarding-rule-v6"
  port_range = "443"

  load_balancing_scheme = "EXTERNAL_MANAGED"
  ip_address            = google_compute_global_address.api_v6.id
  target                = google_compute_target_https_proxy.api.id
}

data "google_dns_managed_zone" "pi_delivery" {
  name = "pi-delivery"
}

resource "google_dns_record_set" "api_a" {
  name         = "api.${data.google_dns_managed_zone.pi_delivery.dns_name}"
  type         = "A"
  ttl          = 300
  managed_zone = data.google_dns_managed_zone.pi_delivery.name
  rrdatas = [
    google_compute_global_address.api.address
  ]
}

resource "google_dns_record_set" "api_aaaa" {
  name         = "api.${data.google_dns_managed_zone.pi_delivery.dns_name}"
  type         = "AAAA"
  ttl          = 300
  managed_zone = data.google_dns_managed_zone.pi_delivery.name
  rrdatas = [
    google_compute_global_address.api_v6.address
  ]
}

resource "google_dns_record_set" "api_staging_a" {
  name         = "api.staging.${data.google_dns_managed_zone.pi_delivery.dns_name}"
  type         = "A"
  ttl          = 60
  managed_zone = data.google_dns_managed_zone.pi_delivery.name
  rrdatas = [
    google_compute_global_address.api.address
  ]
}

resource "google_dns_record_set" "api_staging_aaaa" {
  name         = "api.staging.${data.google_dns_managed_zone.pi_delivery.dns_name}"
  type         = "AAAA"
  ttl          = 60
  managed_zone = data.google_dns_managed_zone.pi_delivery.name
  rrdatas = [
    google_compute_global_address.api_v6.address
  ]
}
