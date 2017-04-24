# GROOTFS BENCH

Benchmarking tool for [grootfs](https://github.com/cloudfoundry/grootfs)

## Usage

```
NAME:
   grootfs-bench - grootfs awesome benchmarking tool

USAGE:
   grootfs-bench --gbin <grootfs-bin> --store <store-path> --log-level <debug|info|warn> --images <n> --concurrency <c> --base-image <docker:///img>

VERSION:
   0.1.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --gbin value                      path to grootfs bin (default: "grootfs")
   --images value                    number of images to create (default: "500")
   --concurrency value               what the name says (default: "5")
   --store value                     store path (default: "/var/lib/grootfs")
   --driver value                    filesystem driver
   --log-level value                 what the name says (default: "debug")
   --base-image value                base image to use
   --with-quota                      add quotas to the image creation
   --nospin                          turn off the awesome spinner, you monster
   --json                            return the result in json format
   --parallel-clean                  run a concurrent clean operation
   --parallel-clean-interval value   interval at which to call clean during concurrent operations in seconds. parallel-clean must also be set (default: 6)
   --parallel-delete-interval value  interval at which to call delete during concurrent operations in seconds. parallel-clean must also be set (default: 3)
   --help, -h                        show help
   --version, -v                     print the version
```

Example:

```
grootfs-bench --gbin /var/vcap/jobs/grootfs/bin/grootfs \
              --store /var/vcap/store/grootfs \
              --images 20 \
              --concurrency 5 \
              --parallel-clean \
              --base-image docker:///ubuntu
              --base-image docker:///alpine
              --base-image docker:///cirros


Total images requested.: 20
Concurrency factor.....: 5
Using Quota?...........: false
Parallel Clean?........: true
Number of cleans.......: 3
Number of deletes......: 6
........................
Total duration.........: 5.696163453s
Images per second......: 3.511135
Average time per image.: 1.405277
Total errors...........: 0
Error Rate.............: 0.000000
```
