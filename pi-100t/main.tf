/**
 * Copyright 2021 Google LLC
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
    bucket = "pi-100t-tf-state"
    prefix = "terraform/state"
  }
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.8.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 4.8.0"
    }
  }
}

// Base images
data "google_compute_image" "windows_base" {
  project = "windows-cloud"
  family  = "windows-2019"
}

data "google_compute_image" "debian_base" {
  project = "debian-cloud"
  family  = "debian-11"
}

// startup script for storage nodes
data "local_file" "startup_script" {
  filename = "${path.module}/startup-script.sh"
}

// Create gVNIC enabled images
resource "google_compute_image" "debian" {
  name         = "debian-gvnic"
  description  = "Debian disk image with gVNIC enabled"
  source_image = data.google_compute_image.debian_base.self_link

  guest_os_features {
    type = "GVNIC"
  }
  guest_os_features {
    type = "UEFI_COMPATIBLE"
  }
  guest_os_features {
    type = "VIRTIO_SCSI_MULTIQUEUE"
  }
}

// Service accounts
data "google_compute_default_service_account" "default" {
}

resource "random_id" "compute_role" {
  byte_length = 8
}

resource "google_service_account" "compute_node" {
  account_id   = "compute-node-sa"
  display_name = "Compute Node SA"
}

resource "google_service_account" "storage_node" {
  account_id   = "storage-node-sa"
  display_name = "Storage Node SA"
}

resource "google_project_iam_custom_role" "snapshot" {
  role_id = "snapshot_role_${random_id.compute_role.hex}"
  title   = "Custom role for the compute node"
  permissions = [
    "compute.disks.createSnapshot",
    "compute.disks.list",
    "compute.disks.get",
    "compute.snapshots.create",
    "compute.snapshots.setLabels",
  ]
}

resource "google_project_iam_binding" "snapshot_compute" {
  project = local.env.project
  role = google_project_iam_custom_role.snapshot.name
  members = [
    "serviceAccount:${google_service_account.compute_node.email}",
  ]
}

resource "google_project_iam_binding" "storage_compute" {
  project = local.env.project
  role = "roles/storage.objectAdmin"
  members = [
    "serviceAccount:${google_service_account.compute_node.email}",
  ]
}

resource "google_project_iam_binding" "compute_viewer" {
  project = local.env.project
  role = "roles/compute.viewer"
  members = [
    "serviceAccount:${google_service_account.compute_node.email}",
    "serviceAccount:${google_service_account.storage_node.email}",
  ]
}

resource "google_project_iam_binding" "storage_object_viewer" {
  project = local.env.project
  role = "roles/storage.objectViewer"
  members = [
    "serviceAccount:${google_service_account.storage_node.email}",
  ]
}

resource "google_project_iam_binding" "logs_writer" {
  project = local.env.project
  role = "roles/logging.logWriter"
  members = [
    "serviceAccount:${google_service_account.compute_node.email}",
    "serviceAccount:${google_service_account.storage_node.email}",
  ]
}

resource "google_project_iam_binding" "monitoring_metric_writer" {
  project = local.env.project
  role = "roles/monitoring.metricWriter"
  members = [
    "serviceAccount:${google_service_account.compute_node.email}",
    "serviceAccount:${google_service_account.storage_node.email}",
  ]
}

resource "google_project_iam_binding" "guest_policy_viewer" {
  project = local.env.project
  role = "roles/osconfig.guestPolicyViewer"
  members = [
    "serviceAccount:${google_service_account.compute_node.email}",
    "serviceAccount:${google_service_account.storage_node.email}",
  ]
}

data "google_compute_network" "default" {
  name = "default"
}

// External addresses
resource "google_compute_address" "compute" {
  name        = "compute-node-ip"
  description = "public IP of the compute node"
}

// Disks to store the final results
// Resize the disks once we get closer to the end.
// Make sure to run mkfs again instead of using xfs_grow because
// there are parameters that don't quite scale from 10 GB to 50 TB.
resource "google_compute_disk" "results_dec" {
  name = "results-dec-disk"
  type = "pd-balanced"
  size = var.result_disk_size
  labels = {
    "snapshot" = "enabled"
    "env"      = local.env.env_label
  }
  lifecycle {
    ignore_changes = [snapshot]
  }
}

resource "google_compute_disk" "results_hex" {
  name = "results-hex-disk"
  type = "pd-balanced"
  size = var.result_disk_size
  labels = {
    "env"      = local.env.env_label
    "snapshot" = "enabled"
  }
  lifecycle {
    ignore_changes = [snapshot]
  }
}

locals {
  total_storage_disk_count = local.env.storage_node_count * local.env.targets_per_node
}

// Disks for y-crunchre swap
resource "google_compute_disk" "storage" {
  count = local.total_storage_disk_count
  name  = "storage-disk-${count.index}"
  type  = local.env.storage_disk_type
  size  = ceil(local.env.total_storage_size * 1000 / local.total_storage_disk_count)
  labels = {
    "snapshot" = "enabled"
    "env"      = local.env.env_label
  }
  lifecycle {
    ignore_changes = [snapshot]
  }
}

// Disk for the y-cruncher directory
resource "google_compute_disk" "y_cruncher" {
  name = "y-cruncher-disk"
  type = "pd-balanced"
  size = 50

  labels = {
    "snapshot" = "enabled"
    "env"      = local.env.env_label
  }
  lifecycle {
    ignore_changes = [snapshot]
  }
}

// Create the compute instance
resource "google_compute_instance" "compute" {
  provider         = google-beta
  name             = "compute-node"
  machine_type     = local.env.compute_node_type
  min_cpu_platform = local.env.compute_cpu_platform

  advanced_machine_features {
    // Disable SMT for better network throughput
    threads_per_core = 1
  }

  boot_disk {
    initialize_params {
      size = 20
      type = "pd-balanced"
      // image = google_compute_image.debian.self_link
      image = data.google_compute_image.debian_gvnic.self_link
      labels = {
        "snapshot" = "enabled"
        "env"      = local.env.env_label
      }
    }
  }

  network_interface {
    network  = data.google_compute_network.default.self_link
    nic_type = "GVNIC"
    access_config {
      nat_ip = google_compute_address.compute.address
    }
  }

  network_performance_config {
    total_egress_bandwidth_tier = "TIER_1"
  }

  shielded_instance_config {
    // enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  attached_disk {
    source      = google_compute_disk.y_cruncher.self_link
    device_name = "y-cruncher"
  }

  attached_disk {
    source      = google_compute_disk.results_dec.self_link
    device_name = "results-dec"
  }

  attached_disk {
    source      = google_compute_disk.results_hex.self_link
    device_name = "results-hex"
  }

  service_account {
    email  = google_service_account.compute_node.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    startup-script = data.local_file.startup_script.content
  }

  labels = {
    type = "compute"
    env  = local.env.env_label
  }

  deletion_protection = local.deletion_protection

  lifecycle {
    ignore_changes = [boot_disk.0.initialize_params]
  }
}

// Create storage nodes for y-cruncher swap
resource "google_compute_instance" "storage" {
  count            = local.env.storage_node_count
  name             = "storage-node-${count.index}"
  machine_type     = local.env.storage_node_type
  min_cpu_platform = local.env.storage_cpu_platform

  advanced_machine_features {
    // Disable SMT for better network throughput
    threads_per_core = 1
  }

  network_interface {
    network  = data.google_compute_network.default.self_link
    nic_type = "VIRTIO_NET"
    access_config {}
  }

  boot_disk {
    initialize_params {
      size = 10
      type = "pd-balanced"
      // image = google_compute_image.debian.self_link
      image = data.google_compute_image.debian_gvnic.self_link
      labels = {
        "env" = local.env.env_label
      }
    }
  }

  dynamic "attached_disk" {
    for_each = slice(google_compute_disk.storage, count.index * local.env.targets_per_node, count.index * local.env.targets_per_node + local.env.targets_per_node)
    iterator = iter
    content {
      source      = iter.value.self_link
      device_name = "storage-disk-${iter.key}"
    }
  }

  shielded_instance_config {
    // enable_secure_boot          = true
    enable_vtpm                 = true
    enable_integrity_monitoring = true
  }

  service_account {
    email  = google_service_account.storage_node.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    startup-script = data.local_file.startup_script.content
  }

  labels = {
    "type" = "storage"
    "env"  = local.env.env_label
  }

  deletion_protection = local.deletion_protection

  lifecycle {
    ignore_changes = [boot_disk.0.initialize_params]
  }
}

// Create a Cloud Storage bucket to copy the final results
resource "google_storage_bucket" "results" {
  name                        = "${local.env.project}-results"
  location                    = "US"
  uniform_bucket_level_access = true
}

// Export the number of nodes as metadata
resource "google_compute_project_metadata_item" "storage_node_count" {
  key   = "storage-node-count"
  value = local.env.storage_node_count
}

resource "google_compute_project_metadata_item" "targets_per_node" {
  key   = "targets-per-node"
  value = local.env.targets_per_node
}

resource "google_compute_project_metadata_item" "snapshot_frequency" {
  key   = "snapshot-frequency"
  value = var.snapshot_frequency
}

resource "google_os_config_guest_policies" "os_agent_debian_11" {
  provider        = google-beta
  guest_policy_id = "ops-agent-debian-11"

  packages {
    name          = "google-cloud-ops-agent"
    desired_state = "UPDATED"
  }

  assignment {
    os_types {
      os_short_name = "debian"
      os_version    = "11"
    }
  }

  package_repositories {
    apt {
      uri          = "https://packages.cloud.google.com/apt"
      archive_type = "DEB"
      distribution = "google-cloud-ops-agent-bullseye-all"
      components   = ["main"]
      gpg_key      = "https://packages.cloud.google.com/apt/doc/apt-key.gpg"
    }
  }
}

resource "google_os_config_guest_policies" "common_packages_deb" {
  provider        = google-beta
  guest_policy_id = "common-packages-deb"

  packages {
    name          = "sysfsutils"
    desired_state = "REMOVED"
  }
  packages {
    name          = "sysstat"
    desired_state = "INSTALLED"
  }

  assignment {
    os_types {
      os_short_name = "debian"
      os_version    = "11"
    }
  }
}

resource "google_os_config_guest_policies" "compute_packages" {
  provider        = google-beta
  guest_policy_id = "compute-packages"

  packages {
    name          = "open-iscsi"
    desired_state = "INSTALLED"
  }

  packages {
    name          = "xfsprogs"
    desired_state = "INSTALLED"
  }

  packages {
    name          = "libnuma1"
    desired_state = "INSTALLED"
  }

  packages {
    name          = "libhugetlbfs-bin"
    desired_state = "INSTALLED"
  }

  packages {
    name          = "numactl"
    desired_state = "INSTALLED"
  }

  assignment {
    os_types {
      os_short_name = "debian"
      os_version    = "11"
    }
    group_labels {
      labels = {
        "type" = "compute"
      }
    }
  }
}
