# Swift health exporter

[![Build Status](https://travis-ci.org/sapcc/swift-health-exporter.svg?branch=master)](https://travis-ci.org/sapcc/swift-health-exporter)
[![Coverage Status](https://coveralls.io/repos/github/sapcc/swift-health-exporter/badge.svg?branch=master)](https://coveralls.io/github/sapcc/swift-health-exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/sapcc/swift-health-exporter)](https://goreportcard.com/report/github.com/sapcc/swift-health-exporter)

This exporter uses the `swift-dispersion-report` and `swift-recon` tools to
emit Prometheus metrics about the health of an OpenStack Swift cluster.

It has been tested with OpenStack Swift Train (v2.23.0) and above. For older
versions of Swift, you might want to look at
[swift-health-statsd](https://github.com/sapcc/swift-health-statsd).

## Installation

The only required build dependency is [Go](https://golang.org/):

```sh
make install
```

This will install the binary in `/usr/bin/`.

## Usage

`swift-health-exporter` will find the executables for `swift-recon` and
`swift-dispersion-report` in the directories named by the `$PATH` environment
variable.

To start collecting metrics, you simply need to:

```sh
swift-health-exporter
```

If the `swift-recon` and `swift-dispersion-report` are not in the directories
named by the `$PATH` then you **must** provide the respective paths to
these executables using the configuration options.

### Configuration options

The following environment variables are recognized:

| Variable                       | Required                          | Description                                                                               |
| ----------------               | ----------------                  | ----------------                                                                          |
| `SWIFT_DISPERSION_REPORT_PATH` | yes, if executable not in `$PATH` | Path to the `swift-dispersion-report` executable.                                         |
| `SWIFT_RECON_PATH`             | yes, if executable not in `$PATH` | Path to the `swift-recon` executable.                                                     |
| `DEBUG`                        | no                                | If this option is set to `true` then `swift-health-exporter` will also output debug logs. |

## Metrics

Metric are exposed at port `9520`. This port has been
[allocated](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)
for `swift-health-exporter`.

### Dispersion Report

| Metric                                       | Labels      |
| ----------------                             | ----------- |
| `swift_dispersion_container_copies_expected` |             |
| `swift_dispersion_container_copies_found`    |             |
| `swift_dispersion_container_copies_missing`  |             |
| `swift_dispersion_container_overlapping`     |             |
| `swift_dispersion_object_copies_expected`    |             |
| `swift_dispersion_object_copies_found`       |             |
| `swift_dispersion_object_copies_missing`     |             |
| `swift_dispersion_object_overlapping`        |             |

### Recon

| Metric                                          | Labels               |
| ----------------                                | -----------          |
| `swift_cluster_accounts_quarantined`            | `storage_ip`         |
| `swift_cluster_accounts_replication_age`        | `storage_ip`         |
| `swift_cluster_accounts_replication_duration`   | `storage_ip`         |
| `swift_cluster_containers_quarantined`          | `storage_ip`         |
| `swift_cluster_containers_replication_age`      | `storage_ip`         |
| `swift_cluster_containers_replication_duration` | `storage_ip`         |
| `swift_cluster_containers_updater_sweep_time`   | `storage_ip`         |
| `swift_cluster_drives_audit_errors`             | `storage_ip`         |
| `swift_cluster_drives_unmounted`                | `storage_ip`         |
| `swift_cluster_md5_${kind}_all`                 |                      |
| `swift_cluster_md5_${kind}_errors`              |                      |
| `swift_cluster_md5_${kind}_matched`             |                      |
| `swift_cluster_md5_${kind}_not_matched`         |                      |
| `swift_cluster_objects_quarantined`             | `storage_ip`         |
| `swift_cluster_objects_replication_age`         | `storage_ip`         |
| `swift_cluster_objects_replication_duration`    | `storage_ip`         |
| `swift_cluster_objects_updater_sweep_time`      | `storage_ip`         |
| `swift_cluster_storage_capacity_bytes`          | `storage_ip`         |
| `swift_cluster_storage_free_bytes`              | `storage_ip`         |
| `swift_cluster_storage_used_bytes`              | `storage_ip`         |
| `swift_cluster_storage_used_percent_by_disk`    | `storage_ip`, `disk` |
| `swift_cluster_storage_used_percent`            | `storage_ip`         |
