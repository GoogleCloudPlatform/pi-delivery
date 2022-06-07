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

while getopts "p:" opt; do
    case $opt in
        p)
            source_project=$OPTARG
            ;;
    esac
done

if [[ -z "$source_project" ]]; then
    source_project=$project
fi

snapshot=($(gcloud compute snapshots list --filter='labels.source=create-snapshot AND name:y-cruncher-disk' --sort-by=~creationTimestamp \
    --limit=1 --format="value(name, creationTimestamp)" --project=$source_project))
snapshot_name=${snapshot[0]}
snapshot_creation_timestamp=${snapshot[1]}
snapshot_suffix=-$(echo "$snapshot_name" | awk -F '-' '{print $4"-"$5}')

echo "The latest snapshot found was created at ${snapshot_creation_timestamp} (with suffix $snapshot_suffix)."

read -p "Are you sure to proceed (yes/no)? " -r
if [ "$REPLY" != "yes" ]; then
    echo "Exiting."
    exit 0
fi

./create-storage-nodes.sh -p "${source_project}" -s "$snapshot_suffix"
./create-compute-node.sh -p "${source_project}" -s "$snapshot_suffix"

storage_node_count=$(terraform output -raw storage_node_count)

read -p "Wait for the storage nodes to become available. Hit enter to continue..."

for ((i=0; i<storage_node_count; i++)); do
    gcloud compute ssh storage-node-$i --zone=$zone --project=$project -- "sudo reboot"
done

echo "Instances recreated. Read README.md for the additional steps to restart y-cruncher."
