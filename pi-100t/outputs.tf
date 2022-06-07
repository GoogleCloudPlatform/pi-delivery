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

output "storage_disk_size" {
  value = google_compute_disk.storage.0.size
}

output "compute_node_ipv4" {
  value = google_compute_address.compute.address
}

output "storage_nodes_ipv4" {
  value = google_compute_instance.storage.*.network_interface.0.access_config.0.nat_ip
}

output "targets_per_node" {
  value = length(google_compute_instance.storage.0.attached_disk)
}

output "storage_node_count" {
  value = length(google_compute_instance.storage)
}

output "storage_node_type" {
  value = google_compute_instance.storage.0.machine_type
}

output "zone" {
  value = google_compute_instance.compute.zone
}

output "project" {
  value = google_compute_instance.compute.project
}