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

locals {
  environments = {
    # default is the primary workspace for the pi-100t project.
    default = {
      # GCP Project ID
      project = "pi-100t"

      # Region to use
      region = "us-central1"

      # Zone for instances and disks
      zone = "us-central1-c"

      # Total storage size in TB
      total_storage_size = 663

      # Number of storage node instances
      storage_node_count = 30

      # Number of iSCSI targets per storage node
      targets_per_node = 2

      # Compute node machine type
      compute_node_type = "n2-highmem-128"

      # Compute node min cpu platform
      compute_cpu_platform = "Intel Ice Lake"

      # Storage node machine type
      storage_node_type = "n2-highcpu-16"

      # Storage node min cpu platform
      storage_cpu_platform = "Intel Ice Lake"

      # Type of disk attached to each storage node
      storage_disk_type = "pd-balanced"

      # Env label applied to each resource
      env_label = "prod"

      # Enable deletion protection for instances
      deletion_protection = true
    }

    # pi_80t is for the pi-80t project (testing/dev).
    pi_80t = {
      project             = "pi-80t"
      region              = "us-central1"
      zone                = "us-central1-c"
      total_storage_size  = 663
      deletion_protection = false
    }
  }
  workspace           = replace(terraform.workspace, "-", "_")
  env                 = merge(local.environments["default"], local.environments[local.workspace])
  deletion_protection = var.deletion_protection == null ? local.env.deletion_protection : var.deletion_protection
}
