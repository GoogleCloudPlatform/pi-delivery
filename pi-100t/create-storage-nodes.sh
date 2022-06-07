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


# You might need to import resources into the terraform state manually.
# Currently, terraform imports the startup script into metadata_startup_script,
# which triggers a replacement when the content is changes.
# This is an undesirable behavior.
# https://github.com/hashicorp/terraform-provider-google/issues/4388
# As a workaround, you can remove the startup script before importing.
#
# ex. (adjust the for loop based on configurations)
# for i in {0..31}; do gcloud compute instances add-metadata storage-node-$i --metadata=startup-script=; done
# for i in {0..63}; do terraform import "google_compute_disk.storage[$i]" storage-disk-$i; done 
# for i in {0..31}; do terraform import "google_compute_instance.storage[$i]" storage-node-$i; done
#
# Make sure to check the status by running terraform plan before making further changes.
# If you mess up, you can always start over by removing the resources from the state file.
#
# terraform state rm google_compute_instance.storage

source ./config.sh

targets_per_node=$(terraform output -raw targets_per_node)
storage_node_count=$(terraform output -raw storage_node_count)
total_disk_count=$(($targets_per_node * $storage_node_count))
storage_disk_size=$(terraform output -raw storage_disk_size)

storage_node_type=$(terraform output -raw storage_node_type)
storage_disk_type=pd-balanced
min_cpu_platform=icelake

while getopts "p:s:" opt; do
    case $opt in
        p)
            source_project=$OPTARG
            ;;
        s)
            snapshot_suffix=$OPTARG
            ;;
    esac
done

if [[ -z "$source_project" ]]; then
    source_project=$project
fi

# create storage nodes
for ((i=0; i<storage_node_count; i++)) do
    create_disk=""
    for ((j=0; j<targets_per_node; j++)) do
        k=$(($i * targets_per_node + $j))
        create_disk="$create_disk --create-disk=name=storage-disk-$k,size=${storage_disk_size}GB,type=${storage_disk_type},device-name=storage-disk-$j,auto-delete=no"
        if [[ $snapshot_suffix ]]; then
            create_disk="$create_disk,source-snapshot=https://compute.googleapis.com/compute/v1/projects/${source_project}/global/snapshots/storage-disk-${k}${snapshot_suffix}"
        fi
    done
    gcloud beta compute instances create storage-node-$i \
        --machine-type=$storage_node_type \
        --min-cpu-platform=$min_cpu_platform \
        --zone=$zone \
        --boot-disk-type=pd-balanced --boot-disk-size=10G --boot-disk-device-name=storage-node-$i\
        --network-interface=network=default,nic-type=GVNIC \
        --shielded-vtpm --shielded-integrity-monitoring \
        $create_disk \
        --image-project=${project} --image=debian-gvnic \
        --service-account=storage-node-sa@${project}.iam.gserviceaccount.com \
        --scopes=cloud-platform \
        --metadata-from-file=startup-script=startup-script.sh \
        --labels=type=storage,env=prod \
        --threads-per-core=1 \
        --project=$project \
        --async
done

wait=1

while [ $wait -ne 0 ]; do
    wait=$(gcloud compute operations list --project=$project --zones=$zone --filter=status=RUNNING --format=list | wc -l)
    sleep 5
done

