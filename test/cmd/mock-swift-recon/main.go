// Copyright 2019 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

func main() {
	serverTypeArg := kingpin.Arg("server-type", "Type of server.").Enum("account", "container", "object")
	_ = kingpin.Flag("timeout", "Time to wait for a response from a server.").Short('t').Int()
	verboseFlag := kingpin.Flag("verbose", "Print verbose info.").Short('v').Bool()
	diskUsageFlag := kingpin.Flag("diskusage", "Get disk usage stats.").Short('d').Bool()
	driveAuditFlag := kingpin.Flag("driveaudit", "Get drive audit error stats.").Bool()
	md5Flag := kingpin.Flag("md5", "Get md5sum of servers ring and compare to local copy.").Bool()
	quarantinedFlag := kingpin.Flag("quarantined", "Get cluster quarantine stats.").Short('q').Bool()
	replicationFlag := kingpin.Flag("replication", "Get replication stats.").Short('r').Bool()
	unmountedFlag := kingpin.Flag("unmounted", "Check cluster for unmounted devices.").Short('u').Bool()
	updaterFlag := kingpin.Flag("updater", "Get updater stats.").Bool()

	kingpin.Parse()
	switch {
	case *diskUsageFlag && *verboseFlag:
		os.Stdout.Write(diskUsageVerboseData)
	case *driveAuditFlag && *verboseFlag:
		os.Stdout.Write(driveAuditVerboseData)
	case *md5Flag:
		os.Stdout.Write(md5Data)
	case *quarantinedFlag && *verboseFlag:
		os.Stdout.Write(quarantinedVerboseData)
	case *replicationFlag && *verboseFlag:
		switch *serverTypeArg {
		case "account":
			os.Stdout.Write(accountReplVerboseData)
		case "container":
			os.Stdout.Write(containerReplVerboseData)
		case "object":
			os.Stdout.Write(objectReplVerboseData)
		}
	case *unmountedFlag && *verboseFlag:
		os.Stdout.Write(unmountedVerboseData)
	case *updaterFlag && *verboseFlag:
		switch *serverTypeArg {
		case "container":
			os.Stdout.Write(containerUpdaterVerboseData)
		case "object":
			os.Stdout.Write(objectUpdaterVerboseData)
		}
	}
}

var diskUsageVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:15:56] Checking disk usage now
-> http://10.0.0.2:6000/recon/diskusage: [{u'device': u'sdb-14', u'avail': 5608006787072, u'mounted': True, u'used': 391029243904, u'size': 5999036030976}, {u'device': u'sdb-03', u'avail': 5621251129344, u'mounted': True, u'used': 377784901632, u'size': 5999036030976}, {u'device': u'sdb-09', u'avail': 5646097563648, u'mounted': True, u'used': 352938467328, u'size': 5999036030976}, {u'device': u'sdb-10', u'avail': 5628910911488, u'mounted': True, u'used': 370125119488, u'size': 5999036030976}, {u'device': u'sdb-06', u'avail': 5615178706944, u'mounted': True, u'used': 383857324032, u'size': 5999036030976}, {u'device': u'sdb-07', u'avail': 5630346199040, u'mounted': True, u'used': 368689831936, u'size': 5999036030976}, {u'device': u'sdb-04', u'avail': 5624401117184, u'mounted': True, u'used': 374634913792, u'size': 5999036030976}, {u'device': u'sdb-05', u'avail': 5632579710976, u'mounted': True, u'used': 366456320000, u'size': 5999036030976}, {u'device': u'sdb-01', u'avail': 5639496806400, u'mounted': True, u'used': 359539224576, u'size': 5999036030976}, {u'device': u'sdb-08', u'avail': 5638742040576, u'mounted': True, u'used': 360293990400, u'size': 5999036030976}, {u'device': u'sdb-12', u'avail': 5635459399680, u'mounted': True, u'used': 363576631296, u'size': 5999036030976}, {u'device': u'sdb-11', u'avail': 5640890974208, u'mounted': True, u'used': 358145056768, u'size': 5999036030976}, {u'device': u'sdb-13', u'avail': 5622460055552, u'mounted': True, u'used': 376575975424, u'size': 5999036030976}, {u'device': u'sdb-02', u'avail': 5599360884736, u'mounted': True, u'used': 399675146240, u'size': 5999036030976}]
-> http://10.0.0.1:6000/recon/diskusage: [{u'device': u'sdb-14', u'avail': 5644861136896, u'mounted': True, u'used': 354174894080, u'size': 5999036030976}, {u'device': u'sdb-03', u'avail': 5635428171776, u'mounted': True, u'used': 363607859200, u'size': 5999036030976}, {u'device': u'sdb-09', u'avail': 5613408104448, u'mounted': True, u'used': 385627926528, u'size': 5999036030976}, {u'device': u'sdb-10', u'avail': 5626034053120, u'mounted': True, u'used': 373001977856, u'size': 5999036030976}, {u'device': u'sdb-06', u'avail': 5627536703488, u'mounted': True, u'used': 371499327488, u'size': 5999036030976}, {u'device': u'sdb-07', u'avail': 5612064350208, u'mounted': True, u'used': 386971680768, u'size': 5999036030976}, {u'device': u'sdb-04', u'avail': 5626924462080, u'mounted': True, u'used': 372111568896, u'size': 5999036030976}, {u'device': u'sdb-05', u'avail': 5619981594624, u'mounted': True, u'used': 379054436352, u'size': 5999036030976}, {u'device': u'sdb-01', u'avail': 5619673370624, u'mounted': True, u'used': 379362660352, u'size': 5999036030976}, {u'device': u'sdb-08', u'avail': 5639914979328, u'mounted': True, u'used': 359121051648, u'size': 5999036030976}, {u'device': u'sdb-12', u'avail': 5615940825088, u'mounted': True, u'used': 383095205888, u'size': 5999036030976}, {u'device': u'sdb-11', u'avail': 5628679016448, u'mounted': True, u'used': 370357014528, u'size': 5999036030976}, {u'device': u'sdb-13', u'avail': 5631554600960, u'mounted': True, u'used': 367481430016, u'size': 5999036030976}, {u'device': u'sdb-02', u'avail': 5626821857280, u'mounted': True, u'used': 372214173696, u'size': 5999036030976}]
Distribution Graph:
  5%    5 ***************
  6%   23 *********************************************************************
Disk usage: space used: 10421003354112 of 167973008867328
Disk usage: space free: 157552005513216 of 167973008867328
Disk usage: lowest: 5.88%, highest: 6.66%, avg: 6.20397492691%
===============================================================================`)

var md5Data = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:14:24] Checking ring md5sums
2/2 hosts matched, 0 error[s] while checking hosts.
===============================================================================
[2019-12-30 00:14:24] Checking swift.conf md5sum
2/2 hosts matched, 0 error[s] while checking hosts.
===============================================================================`)

var containerUpdaterVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (container)
===============================================================================
[2019-12-30 00:12:20] Checking updater times
-> http://10.0.0.1:6001/recon/updater/container: {u'container_updater_sweep': 52.1986780166626}
-> http://10.0.0.2:6001/recon/updater/container: {u'container_updater_sweep': 71.28152513504028}
[updater_last_sweep] low: 52, high: 71, avg: 61.7, total: 123, Failed: 0.0%, no_result: 0, reported: 2
===============================================================================`)

var objectUpdaterVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:12:59] Checking updater times
-> http://10.0.0.2:6000/recon/updater/object: {u'object_updater_sweep': 4.389769792556763}
-> http://10.0.0.1:6000/recon/updater/object: {u'object_updater_sweep': 0.44452810287475586}
[updater_last_sweep] low: 0, high: 4, avg: 2.4, total: 4, Failed: 0.0%, no_result: 0, reported: 2
===============================================================================`)

var accountReplVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (account)
===============================================================================
[2019-12-30 00:11:16] Checking on replication
-> http://10.0.0.1:6002/recon/replication/account: {u'replication_last': 1577664676.578959, u'replication_stats': {u'no_change': 1223, u'rsync': 0, u'success': 1226, u'failure': 0, u'attempted': 613, u'ts_repl': 0, u'remove': 0, u'remote_merge': 0, u'diff_capped': 0, u'deferred': 0, u'hashmatch': 0, u'diff': 3, u'start': 1577664663.576819, u'empty': 0}, u'replication_time': 13.002140045166016}
-> http://10.0.0.2:6002/recon/replication/account: {u'replication_last': 1577664668.9851, u'replication_stats': {u'no_change': 1221, u'rsync': 0, u'success': 1222, u'failure': 0, u'attempted': 611, u'ts_repl': 0, u'remove': 0, u'remote_merge': 0, u'diff_capped': 0, u'deferred': 0, u'hashmatch': 0, u'diff': 1, u'start': 1577664656.767887, u'empty': 0}, u'replication_time': 12.217212915420532}
[replication_failure] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 2
[replication_success] low: 1222, high: 1226, avg: 1224.0, total: 2448, Failed: 0.0%, no_result: 0, reported: 2
[replication_time] low: 12, high: 13, avg: 12.6, total: 25, Failed: 0.0%, no_result: 0, reported: 2
[replication_attempted] low: 611, high: 613, avg: 612.0, total: 1224, Failed: 0.0%, no_result: 0, reported: 2
Oldest completion was 2019-12-30 00:11:08 (7 seconds ago) by 10.0.0.2:6002.
Most recent completion was 2019-12-30 00:11:16 (0 seconds ago) by 10.0.0.1:6002.
===============================================================================`)

var containerReplVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (container)
===============================================================================
[2019-12-30 00:10:09] Checking on replication
-> http://10.0.0.1:6001/recon/replication/container: {u'replication_last': 1577664528.691438, u'replication_stats': {u'no_change': 8298, u'rsync': 0, u'success': 8300, u'failure': 0, u'attempted': 4150, u'ts_repl': 0, u'remove': 0, u'remote_merge': 0, u'diff_capped': 0, u'deferred': 0, u'hashmatch': 0, u'diff': 2, u'start': 1577664444.899301, u'empty': 0}, u'replication_time': 83.79213690757751}
-> http://10.0.0.2:6001/recon/replication/container: {u'replication_last': 1577664555.743305, u'replication_stats': {u'no_change': 8307, u'rsync': 0, u'success': 8308, u'failure': 0, u'attempted': 4154, u'ts_repl': 0, u'remove': 0, u'remote_merge': 0, u'diff_capped': 0, u'deferred': 0, u'hashmatch': 0, u'diff': 1, u'start': 1577664469.557067, u'empty': 0}, u'replication_time': 86.18623805046082}
[replication_failure] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 2
[replication_success] low: 8300, high: 8308, avg: 8304.0, total: 16608, Failed: 0.0%, no_result: 0, reported: 2
[replication_time] low: 83, high: 86, avg: 85.0, total: 169, Failed: 0.0%, no_result: 0, reported: 2
[replication_attempted] low: 4150, high: 4154, avg: 4152.0, total: 8304, Failed: 0.0%, no_result: 0, reported: 2
Oldest completion was 2019-12-30 00:08:48 (1 minutes ago) by 10.0.0.1:6001.
Most recent completion was 2019-12-30 00:09:15 (53 seconds ago) by 10.0.0.2:6001.
===============================================================================`)

var objectReplVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:08:54] Checking on replication
-> http://10.0.0.1:6000/recon/replication/object: {u'replication_last': 1577664310.620143, u'replication_stats': {u'rsync': 9, u'success': 196599, u'failure': 9, u'attempted': 98304, u'remove': 0, u'suffix_count': 1258927, u'start': 1535617859.516976, u'hashmatch': 196599, u'failure_nodes': {u'10.0.0.1': {u'sdb-09': 3}, u'10.0.0.2': {u'sdb-08': 6}}, u'suffix_sync': 0, u'suffix_hash': 2}, u'replication_time': 4.6007425824801125, u'object_replication_last': 1577664310.620143, u'object_replication_time': 4.6007425824801125}
-> http://10.0.0.2:6000/recon/replication/object: {u'replication_last': 1577664316.719913, u'replication_stats': {u'rsync': 12, u'success': 196596, u'failure': 12, u'attempted': 98304, u'remove': 0, u'suffix_count': 1258168, u'start': 1535618096.714857, u'hashmatch': 196596, u'failure_nodes': {u'10.0.0.1': {u'sdb-13': 1, u'sdb-01': 1, u'sdb-02': 1, u'sdb-09': 4}, u'10.0.0.2': {u'sdb-06': 1, u'sdb-13': 1, u'sdb-08': 3}}, u'suffix_sync': 0, u'suffix_hash': 1}, u'replication_time': 4.947240881125132, u'object_replication_last': 1577664316.719913, u'object_replication_time': 4.947240881125132}
[replication_failure] low: 9, high: 12, avg: 10.5, total: 21, Failed: 0.0%, no_result: 0, reported: 2
[replication_success] low: 196596, high: 196599, avg: 196597.5, total: 393195, Failed: 0.0%, no_result: 0, reported: 2
[replication_time] low: 4, high: 4, avg: 4.8, total: 9, Failed: 0.0%, no_result: 0, reported: 2
[replication_attempted] low: 98304, high: 98304, avg: 98304.0, total: 196608, Failed: 0.0%, no_result: 0, reported: 2
Oldest completion was 2019-12-30 00:05:10 (3 minutes ago) by 10.0.0.1:6000.
Most recent completion was 2019-12-30 00:05:16 (3 minutes ago) by 10.0.0.2:6000.
===============================================================================`)

var quarantinedVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:07:57] Checking quarantine
-> http://10.0.0.2:6000/recon/quarantined: {u'objects': 0, u'accounts': 0, u'containers': 0, u'policies': {}}
-> http://10.0.0.1:6000/recon/quarantined: {u'objects': 0, u'accounts': 0, u'containers': 0, u'policies': {}}
[quarantined_objects] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 2
[quarantined_accounts] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 2
[quarantined_containers] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 2
===============================================================================`)

var unmountedVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:06:50] Getting unmounted drives from 2 hosts...
-> http://10.0.0.1:6000/recon/unmounted: []
-> http://10.0.0.2:6000/recon/unmounted: [{u'device': u'sdb-01', u'mounted': False}]
Not mounted: sdb-01 on 10.0.0.2:6000
===============================================================================`)

var driveAuditVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2019-12-30 00:05:13] Checking drive-audit errors
-> http://10.0.0.1:6000/recon/driveaudit: {u'drive_audit_errors': 0}
-> http://10.0.0.2:6000/recon/driveaudit: {u'drive_audit_errors': 1}
[drive_audit_errors] low: 0, high: 1, avg: 0.5, total: 0, Failed: 0.0%, no_result: 0, reported: 2
===============================================================================`)
