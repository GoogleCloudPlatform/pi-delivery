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


set -o pipefail

now=$(date --rfc-3339=ns)
# unit is minutes
snapshot_frequency=$(curl -f -s -H "Metadata-Flavor:Google" http://metadata/computeMetadata/v1/project/attributes/snapshot-frequency)
script_dir="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
lock_file="$script_dir/last-snapshot.lock"

log() {
    local prio="$1"
    shift
    local s=${*:-$(</dev/stdin)}
    logger --id="$$" --priority "user.$prio" $s
}

log info "snapshot frequency: $snapshot_frequency minutes"

if [[ -f "$lock_file" ]] && [[ $(find "$lock_file" -nowarn -newermt "$(date --date="$now -$snapshot_frequency minutes" --rfc-3339=ns)" -print) ]]; then
    log info "lock file last updated at $(stat --format %y "$lock_file"), exiting"
    exit
fi

sync
find /mnt -nowarn -maxdepth 1 -type d -print0 | xargs -P16 -l -0 -- fstrim -v | log info
sync

zone=$(curl -f -s -H "Metadata-Flavor:Google" http://metadata.google.internal/computeMetadata/v1/instance/zone | cut -d '/' -f 4)
disks=($(gcloud compute disks list --filter=labels.snapshot=enabled --zones=$zone --format="value(name)"))

snapshot_prefix=
snapshot_suffix=-$(date --date="$now" +%Y%m%d-%H%M%S)
snapshots=()

for disk in "${disks[@]}"; do
    snapshots+=("$snapshot_prefix$disk$snapshot_suffix")
done

snapshot_names=$(IFS=, ; echo "${snapshots[*]}")

gcloud compute disks snapshot "${disks[@]}" --snapshot-names="${snapshot_names}" \
    --zone=$zone --labels=source=create-snapshot --async |& log info

if [[ $? -ne 0 ]]; then
    log err "gcloud command failed to create snapshots"
    exit
fi

touch --date="$now" "$lock_file"
