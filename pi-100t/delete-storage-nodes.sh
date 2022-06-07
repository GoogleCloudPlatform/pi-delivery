#!/bin/bash
# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


source ./config.sh

targets_per_node=$(terraform output -raw targets_per_node)
storage_node_count=$(terraform output -raw storage_node_count)
total_disk_count=$(($targets_per_node * $storage_node_count))

declare nodes

for ((i=0; i<storage_node_count; i++)); do
    nodes="$nodes storage-node-$i"
done

echo gcloud compute instances delete $nodes --zone=$zone --project=$project

declare disks

for ((i=0; i<total_disk_count; i++)) do
    disks="$disks storage-disk-$i"
done

echo gcloud compute disks delete $disks --zone=$zone --project=$project
