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

Upload the raw image in a bucket and create an instance:

```
cd tamago-example
make example TARGET=microvm img
cp example.img disk.raw
tar --format=oldgnu -Sczf compressed-image.tar.gz disk.raw
gcloud storage buckets create gs://tamago-bucket
gcloud storage cp compressed-image.tar.gz gs://tamago-bucket
gcloud compute images create tamago-example --source-uri gs://tamago-bucket/compressed-image.tar.gz --architecture=X86_64
gcloud compute instances create tamago-example --zone=europe-west1-b --machine-type=n1-standard-2 --metadata="serial-port-enable=true" --image tamago-example
```

Check the serial port output:

```
gcloud compute instances get-serial-port-output tamago-example --zone=europe-west1-b --port=1
```

Clean up:

```
gcloud storage rm gs://tamago-bucket/compressed-image.tar.gz
gcloud storage buckets delete gs://tamago-bucket
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
