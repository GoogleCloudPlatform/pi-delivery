# Pi 100t experiments

This folder contains scripts that we used to calculate [100 trillion digits of pi](https://cloud.google.com/blog/products/compute/calculating-100-trillion-digits-of-pi-on-google-cloud).
It contains project-specific parameters (such as project ID) so it won't work as it is, but we plan to publish more content explaining techniques and decision decisions in the future.

# First setup

Set the default project and zone for the gcloud command.

```bash
gcloud config set project pi-100t
gcloud config set compute/zone us-central1-a
```

Run `terraform init && terraform apply` and terraform will take care of the initial setup. 

## Reboot the machines at least once

The sysctl and kernel parameter changes are not reflected until you reboot the machines at least once. Make sure to unmount / disconnect iSCSI targets, and then reboot the storage nodes first. Compute node should be rebooted after they become available.

```bash
# Change 29 to (the number of storage nodes - 1)
for i in {0..29}; do gcloud compute ssh storage-node-$i --zone=us-central1-c  -- sudo reboot; done
# wait for a few minutes
gcloud compute ssh compute-node --zone=us-central1-c -- sudo reboot
```

## Mount disks

Run `init-disks.sh` on the compute node to mount the disks. Also run `add-result-disks.sh` to prepare the placeholders for the result disks.


## Manual files

Some files need manual handling.

Copy the files under `etc/` to the `/etc` directory of the compute node.

Set up y-cruncher.

```bash
sudo -s # Become root
cd /mnt/y-cruncher
gsutil cp 'gs://pi-100t-internal/y-cruncher%20v0.7.8.9507-dynamic.tar.xz' ./
tar xf y-cruncher%20v0.7.8.9507-dynamic.tar.xz -C y-cruncher --strip-components 1
gsutil cp gs://pi-100t-internal/100t-60.cfg ./y-cruncher/
```

Copy `create-snapshot.sh` to the same directory as y-cruncher (`/mnt/y-cruncher/y-cruncher`) and make it executable.

Now you can run y-cruncher by running

```bash
cd /mnt/y-cruncher/y-cruncher
./y-cruncher
```

IMPORTANT: Run y-cruncher inside `sudo tmux` so that it can keep running when the ssh is disconnected. You can reattach to the session by running `sudo tmux attach`.

When everything looks fine, protect the instances from deletion.

```bash
terraform apply -var "deletion_protection=true"
```

## Snapshots

y-cruncher lets you set up a post-checkpoint script. `create-snapshot.sh` was created to take advantage of this feature and creates PD snapshots
every `snapshot_frequency` minutes (see `variables.tf`). The script creates `last-snapshot.lock` in the same directory.
Deleting this file resets the timer and forces the command to take snapshots when invoked next time.

You can delete old snapshots with `delete-snapshots.sh`. The command deletes snapshots older than 30 days with a `source=create-snapshot` label by default. You can change the window by passing an argument in the [ISO 8601 duration format](https://en.wikipedia.org/wiki/ISO_8601#Durations). For example, the following command deletes snapshots created more than 15 minutes ago.

```bash
./delete-snapshot.sh PT15M
```

# Restarting from checkpoints

Warning: Take snapshots before trying anything destructive.

Take snapshots manually by running
```bash
cd /mnt/y-cruncher/y-cruncher
rm last-snapshot.lock
bash ./create-snapshot.sh
```

y-cruncher saves checkpoints to the swap disks and you can restart from them. First, make sure you have checkpoint files in /mnt/disk*.

```bash
find /mnt -path '/mnt/disk*' -name '*checkpoint*' -type f
```

This should output checkpoint files. You can delete the other files safely to save storage space.

```bash
find /mnt -path '/mnt/disk*' ! -name '*checkpoint*' -type f
# check the output
find /mnt -path '/mnt/disk*' ! -name '*checkpoint*' -type f -delete
```

Now you can just restart y-cruncher by running the following command inside a `tmux` session.
```
./y-cruncher
```

# Restarting from snapshots

Things happen. Here's how to restart from the snapshots.

If you can log in to the compute node and want to reuse it, unmount the disks.

```bash
umount /mnt/disk*
umount /mnt/y-cruncher
iscsiadm -m node -u
iscsiadm -m node -o delete

# Make sure all nodes are deleted.
iscsiadm -m node
```

Delete the storage instances and disks by running the commands shown by running the following command.

```bash
terraform apply -var "deletion_protection=false"
./delete-storage-nodes.sh
```

If you think the compute node needs to be recreated, delete it as well (optional):

```bash
./delete-compute-node.sh
```

Then run

```bash
./restart-from-snapshots.sh
```

You will see an 'already exists' error if you didn't delete the compute node. It is safe to ignore it.

Now you have the machines running. Run `terraform apply` to apply any changes that are missing. There is some manual work. Log in to the compute node and run the following commands as root.

IMPORTANT: Run y-cruncher inside `sudo tmux` so that it can keep running when the ssh is disconnected.

```bash
umount /mnt/disk*
iscsiadm -m node -u
iscsiadm -m node -o delete

# Make sure all nodes are deleted.
iscsiadm -m node

storage_node_count=$(curl -f -s -H "Metadata-Flavor:Google" http://metadata/computeMetadata/v1/project/attributes/storage-node-count)
for ((i=0; i<storage_node_count; i++)) do
    sudo iscsiadm -m discovery --op=new --type st --portal "storage-node-$i"
done

sudo iscsiadm -m node --loginall=automatic
mount -a

cd /mnt/y-cruncher/y-cruncher
./y-cruncher
```

y-cruncher should automatically restart from the checkpoint.

Make sure to protect the instances by running:
```bash
terraform apply -var "deletion_protection=true"
```

# Resizing result disks

The result disks are created as 10GB disks at first to reduce waste. When we approach the end of the calculation (when y-cruncher finishes the series is probably a good timing), we need to resize the disks. First, resize them by running

```bash
terraform apply -var 'result_disk_size=50000'
```

50 TB should be good for 100 trillion digits, and 40 TB for 80 trillion.

Log in to the compute node and run the following commands as root.

```bash
umount /mnt/results-dec
umount /mnt/results-hex

eval $(grep results-dec /etc/fstab | awk '{print $1}')
mkfs.xfs -f -m uuid=$UUID /dev/disk/by-id/google-results-dec

eval $(grep results-hex /etc/fstab | awk '{print $1}')
mkfs.xfs -f -m uuid=$UUID /dev/disk/by-id/google-results-hex

mount /mnt/results-dec
mount /mnt/results-hex
mkdir '/mnt/results-dec/Pi - Dec - Chudnovsky'
mkdir '/mnt/results-hex/Pi - Hex - Chudnovsky'
```

Confirm the disks are resized by running
```bash
df -h /mnt/results-dec /mnt/results-hex
```
