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

machine_image=--source-machine-image=compute-node-image

while getopts "p:s:" opt; do
    case $opt in
        p)
            source_project=$OPTARG
            ;;
        s)
            machine_image=
            snapshot_suffix=$OPTARG
            ;;
    esac
done

if [[ -z "$source_project" ]]; then
    source_project=$project
fi

function create_disk {
    local name=$1
    local device_name=$2
    local snapshot_suffix=$3
    local s="--create-disk=name=$name,device-name=$device_name,type=pd-balanced"
    if [[ $snapshot_suffix ]]; then
        s="${s},source-snapshot=https://compute.googleapis.com/compute/v1/projects/${source_project}/global/snapshots/${name}${snapshot_suffix}"
    fi
    echo "$s"
}

# create instance from either the machine image or 
echo gcloud beta compute instances create compute-node \
    $machine_image \
    --machine-type=n2-highmem-128 \
    --min-cpu-platform=icelake \
    --zone=$zone \
    --network-interface=network=default,nic-type=GVNIC,address=$(terraform output -raw compute_node_ipv4) \
    --network-performance-configs=total-egress-bandwidth-tier=TIER_1 \
    "$(create_disk compute-node-disk persistent-disk-0 $snapshot_suffix),size=20GB,boot=yes" \
    "$(create_disk y-cruncher-disk y-cruncher $snapshot_suffix),auto-delete=no" \
    "$(create_disk results-dec-disk results-dec $snapshot_suffix),auto-delete=no" \
    "$(create_disk results-hex-disk results-hex $snapshot_suffix),auto-delete=no" \
    --shielded-vtpm --shielded-integrity-monitoring \
    --service-account=compute-node-sa@${project}.iam.gserviceaccount.com \
    --scopes=cloud-platform \
    --metadata-from-file=startup-script=startup-script.sh \
    --labels=type=compute,env=prod \
    --threads-per-core=1 \
    --project=$project

exit

gcloud compute disks add-labels compute-node-disk --labels=env=prod,snapshot=enabled --zone=$zone --project=$project
gcloud compute disks add-labels y-cruncher-disk --labels=env=prod,snapshot=enabled --zone=$zone --project=$project
gcloud compute disks add-labels results-dec-disk --labels=env=prod,snapshot=enabled --zone=$zone --project=$project
gcloud compute disks add-labels results-hex-disk --labels=env=prod,snapshot=enabled --zone=$zone --project=$project
