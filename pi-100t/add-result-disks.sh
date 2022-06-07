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


sudo mkfs.xfs /dev/disk/by-id/google-results-dec
sudo mkfs.xfs /dev/disk/by-id/google-results-hex
sudo mkdir -p /mnt/results-{dec,hex}

sudo blkid /dev/disk/by-id/google-results-dec |awk '{print $2" /mnt/results-dec    xfs    defaults   0 2"}' | sudo tee -a /etc/fstab
sudo blkid /dev/disk/by-id/google-results-hex |awk '{print $2" /mnt/results-hex    xfs    defaults   0 2"}' | sudo tee -a /etc/fstab

mount /mnt/results-dec
mount /mnt/results-hex

sudo mkdir '/mnt/results-dec/Pi - Dec - Chudnovsky'
sudo mkdir '/mnt/results-hex/Pi - Hex - Chudnovsky'

sudo ln -s '/mnt/results-dec/Pi - Dec - Chudnovsky' /mnt/y-cruncher/results/
sudo ln -s '/mnt/results-hex/Pi - Hex - Chudnovsky' /mnt/y-cruncher/results/
