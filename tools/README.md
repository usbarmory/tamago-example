This directory contains helpers for execution of `GOOS=tamago` unikernels on
various platforms.

Boot disk image for cloud deployments
=====================================

The `build_mbr.sh` scripts aids creation of raw boot disk images suitable for
execution of tamago/amd64 unikernels on cloud VM environments.

Google Compute Engine
---------------------

The following example adapts a tamago/amd64 unikernel for execution on Google
Compute Engine, deploying with Google Cloud CLI (though tools like
[ops](https://github.com/nanovms/ops) can also be used).

First of all a unique bucket name should be picked:

```
export BUCKET=tamago-example-$(date +%s)-bucket
```

Upload the raw image in a bucket and create an instance:

```
make example TARGET=gcp img
cp example.img disk.raw
tar --format=oldgnu -Sczf compressed-image.tar.gz disk.raw
gcloud storage buckets create gs://$BUCKET
gcloud storage cp compressed-image.tar.gz gs://$BUCKET
gcloud compute images create tamago-example --source-uri gs://$BUCKET/compressed-image.tar.gz --architecture=X86_64
gcloud compute instances create tamago-example --zone=europe-west1-b --machine-type=t2d-standard-1 --metadata="serial-port-enable=1" --image tamago-example --private-network-ip 10.132.0.2
```

Connect to serial port:

```
# output: unikernels with serial output only, no console input
gcloud compute instances get-serial-port-output tamago-example --zone=europe-west1-b --port=1

# input/output: unikernels with interactive serial console
gcloud compute connect-to-serial-port tamago-example --zone=europe-west1-b --port=1
```

Stop and delete instance:

```
gcloud storage rm gs://$BUCKET/compressed-image.tar.gz
gcloud storage buckets delete gs://$BUCKET
gcloud compute instances stop tamago-example --zone europe-west1-b
gcloud compute instances delete tamago-example --zone europe-west1-b
gcloud compute images delete tamago-example
```

BIOS for QEMU riscv64
=====================

The `build_riscv64_bios.sh` script aids creation of a first stage binary
suitable for `qemu-system-riscv64 -bios` flag.

Example use can be found in `TARGET=sifive_u` in
[tamago-example Makefile](https://github.com/usbarmory/tamago-example/blob/master/Makefile).

  * to launch `tamago/riscv64` binaries under qemu: bios.
  * to launch `tamago/amd64` binaries as disk images under qemu and supported
    cloud KVMs: mbr.
