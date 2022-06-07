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

# Common configs for compute and storage nodes (for now)

# Disable security features
systemctl stop apparmor
systemctl disable apparmor
echo 'GRUB_CMDLINE_LINUX_DEFAULT="$GRUB_CMDLINE_LINUX_DEFAULT apparmor=0 mitigations=off"' \
    | tee /etc/default/grub.d/unsecure.cfg

cat <<EOF > /etc/udev/rules.d/99-local.rules
ACTION!="add|change", GOTO="rules_end"

KERNEL=="sd*", SUBSYSTEM=="block", ATTR{queue/scheduler}="none"
SUBSYSTEM=="scsi", ATTRS{vendor}=="LIO-ORG*", ATTR{timeout}="1800"

LABEL="rules_end"
EOF

update-grub

cat <<EOF > /etc/sysctl.d/local.conf
net.core.rmem_max=536870912
net.core.wmem_max=536870912
net.ipv4.tcp_rmem=4096 131072 536870912
net.ipv4.tcp_wmem=4096 16384 536870912
net.ipv4.tcp_mtu_probing=1

net.core.default_qdisc=fq
# net.core.default_qdisc=pfifo_fast
net.ipv4.tcp_congestion_control=bbr
# net.ipv4.tcp_congestion_control=cubic

vm.swappiness=0
EOF

# The default 1024 is too small.
cat <<EOF > /etc/security/limits.d/local.conf
root            soft     nofile          1048576
*               soft     nofile          1048576
EOF

# Prevent the heavy part from running every time

lock_file=/root/init_script_lock
if [[ -f $lock_file ]]; then
    exit
fi

hostname=$(hostname)

if [[ $hostname == *"compute"* ]]; then
    exit
fi

# The rest is specifically for storage nodes

exit_code=1
retry=20

while [ $exit_code -ne 0 ]; do
    retry=$((retry-1))
    if [ $retry -lt 0 ]; then
        echo "timed out while waiting for apt lock"
    fi

    DEBIAN_FRONTEND=noninteractive apt-get -y update && apt-get -y install targetcli-fb
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        break
    fi
    sleep 10
done

# Set up iscsi

iqn_base=iqn.2003-01.org.linux-iscsi.${hostname}
initiator=iqn.1993-08.org.debian:compute-node
disk_label_base=/dev/disk/by-id/google-storage-disk

set -e

set_parameter () {
    targetcli "iscsi/$1/tpg1/" set parameter "$2"
}

targets_per_node=$(curl -f -s -H "Metadata-Flavor:Google" http://metadata/computeMetadata/v1/project/attributes/targets-per-node)

for ((i=0; i<targets_per_node; i++)) do
    block_storage="storage-disk-${i}"
    iqn="${iqn_base}:disk${i}"

    targetcli backstores/block/ create "${block_storage}" "${disk_label_base}-${i}"
    targetcli "backstores/block/${block_storage}" set attribute emulate_tpu=1
    targetcli iscsi/ create "${iqn}"
    targetcli "iscsi/${iqn}/tpg1/luns/" create "/backstores/block/${block_storage}"
    targetcli "iscsi/${iqn}/tpg1/acls/" create "${initiator}"
    set_parameter "${iqn}" InitialR2T=No
    set_parameter "${iqn}" MaxBurstLength=16776192
    set_parameter "${iqn}" FirstBurstLength=262144
    set_parameter "${iqn}" MaxRecvDataSegmentLength=262144
    set_parameter "${iqn}" MaxXmitDataSegmentLength=262144
    set_parameter "${iqn}" MaxOutstandingR2T=1
done

targetcli saveconfig

touch $lock_file
