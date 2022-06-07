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


variable "deletion_protection" {
  description = "Override deletion protection for instances"
  type        = bool
  default     = null
}

variable "snapshot_frequency" {
  description = "How often we take snapshots of y-cruncher checkpoint files in minutes"
  type        = number
  default     = 3 * 24 * 60 // once in three days
}

variable "result_disk_size" {
  description = "Size of the result disks in GB"
  type        = number
  default     = 10
}
