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


param="$1"

if [ -z "$param" ]; then
    echo "usage: $0 param"
    echo "e.g. $0 MaxOutstandingR2T=32"
    exit
fi

targets_per_node=$(terraform output -raw targets_per_node)
storage_node_count=$(terraform output -raw storage_node_count)

echo "Setting ${param} across $storage_node_count nodes ($targets_per_node per node)"

for ((i=0; i<storage_node_count; i++)); do
    gcloud compute ssh storage-node-$i --zone=us-central1-a -- \
    'for i in {0..'$(($targets_per_node-1))'}; do sudo targetcli iscsi/iqn.2003-01.org.linux-iscsi.$(hostname):disk$i/tpg1/ set parameter '"${param}"' ; done && sudo targetcli saveconfig'
done
