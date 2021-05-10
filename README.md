# Swift health exporter

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/sapcc/swift-health-exporter/Build%20and%20Test)](https://github.com/sapcc/swift-health-exporter/actions?query=workflow%3A%22Build+and+Test%22)
[![Coverage Status](https://img.shields.io/coveralls/github/sapcc/swift-health-exporter)](https://coveralls.io/github/sapcc/swift-health-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/sapcc/swift-health-exporter)](https://goreportcard.com/report/github.com/sapcc/swift-health-exporter)

This exporter uses the `swift-dispersion-report` and `swift-recon` tools to
emit Prometheus metrics about the health of an OpenStack Swift cluster.

The exporter has been tested with OpenStack Swift Train (v2.23.0) and above.
For older versions of Swift, you might want to look at
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

Metric are exposed at port `9520`. This port has been
[allocated](https://github.com/prometheus/prometheus/wiki/Default-port-allocations)
for `swift-health-exporter`.

If the `swift-recon` and `swift-dispersion-report` are not in the directories
named by the `$PATH` then you **must** provide the respective paths to
these executables using the configuration options.

### Configuration options

The following environment variables are recognized:

| Variable                       | Required                                                                      | Description                                                                               |
| ----------------               | ----------------                                                              | ----------------                                                                          |
| `SWIFT_DISPERSION_REPORT_PATH` | yes, if executable not in `$PATH` and `dispersion` collector is enabled       | Path to the `swift-dispersion-report` executable.                                         |
| `SWIFT_RECON_PATH`             | yes, if executable not in `$PATH` and any `recon.<name>` collector is enabled | Path to the `swift-recon` executable.                                                     |
| `DEBUG`                        | no                                                                            | If this option is set to `true` then `swift-health-exporter` will also output debug logs. |

## Collectors

Collectors are enabled by providing a `--collector.<name>` flag. Collectors
that are enabled by default can be disabled by providing a
`--no-collector.<name>` flag.

| Name                       | Enabled by default |
| -------                    | -------            |
| `dispersion`               | no                 |
| `recon.diskusage`          | no                 |
| `recon.driveaudit`         | no                 |
| `recon.md5`                | yes                |
| `recon.quarantined`        | no                 |
| `recon.replication`        | no                 |
| `recon.unmounted`          | no                 |
| `recon.updater_sweep_time` | no                 |

Optionally host timeout for recon collector and context timeout for both
collectors can be provided using the respective flags. Use `--help` for usage
info and default timeout values.

## Metrics

### dispersion

| Metric                                       | Labels  |
| -------                                      | ------- |
| `swift_dispersion_container_copies_expected` |         |
| `swift_dispersion_container_copies_found`    |         |
| `swift_dispersion_container_copies_missing`  |         |
| `swift_dispersion_container_overlapping`     |         |
| `swift_dispersion_object_copies_expected`    |         |
| `swift_dispersion_object_copies_found`       |         |
| `swift_dispersion_object_copies_missing`     |         |
| `swift_dispersion_object_overlapping`        |         |
| `swift_dispersion_task_exit_code`            | `query` |
| `swift_dispersion_errors`                    |         |

### recon

| Metric                       | Labels  |
| -------                      | ------- |
| `swift_recon_task_exit_code` | `query` |


#### recon.diskusage

| Metric                                       | Labels               |
| ----------------                             | -----------          |
| `swift_cluster_storage_capacity_bytes`       |                      |
| `swift_cluster_storage_free_bytes`           |                      |
| `swift_cluster_storage_used_bytes`           |                      |
| `swift_cluster_storage_used_percent_by_disk` | `storage_ip`, `disk` |
| `swift_cluster_storage_used_percent`         |                      |

#### recon.driveaudit

| Metric                              | Labels       |
| -------                             | -------      |
| `swift_cluster_drives_audit_errors` | `storage_ip` |

#### recon.md5

| Metric                          | Labels               |
| -------                         | -------              |
| `swift_cluster_md5_all`         | `kind`               |
| `swift_cluster_md5_errors`      | `storage_ip`, `kind` |
| `swift_cluster_md5_matched`     | `storage_ip`, `kind` |
| `swift_cluster_md5_not_matched` | `storage_ip`, `kind` |

#### recon.quarantined

| Metric                                 | Labels       |
| -------                                | -------      |
| `swift_cluster_accounts_quarantined`   | `storage_ip` |
| `swift_cluster_containers_quarantined` | `storage_ip` |
| `swift_cluster_objects_quarantined`    | `storage_ip` |

#### recon.replication

| Metric                                          | Labels       |
| -------                                         | -------      |
| `swift_cluster_accounts_replication_age`        | `storage_ip` |
| `swift_cluster_accounts_replication_duration`   | `storage_ip` |
| `swift_cluster_containers_replication_age`      | `storage_ip` |
| `swift_cluster_containers_replication_duration` | `storage_ip` |
| `swift_cluster_objects_replication_age`         | `storage_ip` |
| `swift_cluster_objects_replication_duration`    | `storage_ip` |

#### recon.unmounted

| Metric                           | Labels       |
| -------                          | -------      |
| `swift_cluster_drives_unmounted` | `storage_ip` |

#### recon.updater_sweep_time

| Metric                                        | Labels       |
| -------                                       | -------      |
| `swift_cluster_containers_updater_sweep_time` | `storage_ip` |
| `swift_cluster_objects_updater_sweep_time`    | `storage_ip` |
