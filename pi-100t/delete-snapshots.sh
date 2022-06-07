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

# default: delete snapshots created more than 30 days ago.
cutoff=${1:-"P30D"}
filter="labels.source=create-snapshot"

remaining=$(gcloud compute snapshots list --project=$project --filter="$filter AND creationTimestamp>=-$cutoff" --limit=10 --format="value(name)")

if [ -z "$remaining" ]; then
    echo "This would delete all snapshots. Exiting."
    exit
fi

echo "These snapshots will remain (first 10):"
echo "$remaining"

delete=($(gcloud compute snapshots list --project=$project --filter="$filter AND creationTimestamp<-$cutoff" --format="value(name)"))

if [ -z "$delete" ]; then
    echo "No snapshots to delete. Exiting."
    exit
fi

echo "The following snapshots will be deleted:"
echo "${delete[@]}"

read -p "${#delete[@]} snapshots will be deleted. Are you sure to proceed (yes/no)? " -r
if [ "$REPLY" != "yes" ]; then
    echo "Exiting."
    exit 0
fi

gcloud compute snapshots delete --quiet --project=$project "${delete[@]}"
