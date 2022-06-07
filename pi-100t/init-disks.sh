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

sudo apt -y update && sudo apt -y install xfsprogs libnuma1 open-iscsi libhugetlbfs-bin

storage_node_count=$(curl -f -s -H "Metadata-Flavor:Google" http://metadata/computeMetadata/v1/project/attributes/storage-node-count)
targets_per_node=$(curl -f -s -H "Metadata-Flavor:Google" http://metadata/computeMetadata/v1/project/attributes/targets-per-node)
sudo sed -i -E 's/(InitiatorName=).*/\1iqn.1993-08.org.debian:compute-node/' /etc/iscsi/initiatorname.iscsi
sudo systemctl restart iscsid open-iscsi

# Assuming zz is enough (698 devices)
chars=( {e..z} {a..z}{a..z} )

for ((i=0; i<storage_node_count; i++)) do
    sudo iscsiadm -m discovery --op=new --type st --portal "storage-node-$i"
done

sudo iscsiadm -m node --loginall=automatic

sudo cp /etc/fstab /etc/fstab.orig

total_disk_count=$(($storage_node_count * $targets_per_node))

for ((i=0; i<total_disk_count; i++)) do echo /dev/sd${chars[$i]}; done \
    | xargs -P 16 -l sudo mkfs.xfs -f
    # sudo mkfs.ext4 -F -E lazy_itable_init=0,lazy_journal_init=0,discard

for ((i=0; i<total_disk_count; i++)) do
    sudo mkdir -p /mnt/disk$i
    sudo blkid /dev/sd${chars[$i]} |awk "{print \$2\" /mnt/disk$i    xfs    defaults,noatime,_netdev   0 2\"}" | sudo tee -a /etc/fstab
    # sudo blkid /dev/sd${chars[$i]} |awk "{print \$2\" /mnt/disk$i    ext4    defaults,noatime,_netdev   0 2\"}" | sudo tee -a /etc/fstab
done

sudo mkfs.xfs /dev/disk/by-id/google-y-cruncher
sudo mkdir -p /mnt/y-cruncher
sudo blkid /dev/disk/by-id/google-y-cruncher |awk '{print $2" /mnt/y-cruncher    xfs    defaults   0 2"}' | sudo tee -a /etc/fstab

sudo mount -a
sudo systemctl restart sysfsutils
