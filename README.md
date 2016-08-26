# GROOTFS BENCH

Benchmarking tool for [grootfs](https://github.com/cloudfoundry/grootfs)

## Usage

```
NAME:
   grootfs-bench - grootfs awesome benchmarking tool

USAGE:
   grootfs-bench --gbin <grootfs-bin> --store <btrfs-store> --bundles <n> --concurrency <c> --image <docker:///img>

VERSION:
   0.1.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --gbin value         path to grootfs bin (default: "grootfs")
   --bundles value      number of bundles to create (default: "1000")
   --concurrency value  what the name says (default: "5")
   --store value        store path (default: "/var/lib/grootfs")
   --image value        image to use (default: "docker:///busybox:latest")
   --help, -h           show help
   --version, -v        print the version
```

Example:

```
grootfs-bench --gbin /var/vcap/jobs/grootfs/bin/grootfs \
              --store /var/vcap/store/grootfs \
              --bundles 20 \
              --concurrency 5 \
              --image docker:///ubuntu


........................
Total duration.........: 5.696163453s
Bundles per second.....: 3.511135
Average time per bundle: 1.405277
Total errors...........: 0
Error Rate.............: 0.000000
```
