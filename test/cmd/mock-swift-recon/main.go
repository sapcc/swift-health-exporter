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

	"github.com/alecthomas/kong"
)

var cli struct {
	Timeout     int    `short:"t" help:"Time to wait for a response from a server."`
	ServerType  string `arg:"" optional:"" help:"Type of server."`
	Verbose     bool   `short:"v" help:"Print verbose info."`
	Diskusage   bool   `short:"d" name:"diskusage" help:"Get disk usage stats."`
	Driveaudit  bool   `name:"driveaudit" help:"Get drive audit error stats."`
	MD5         bool   `name:"md5" help:"Get md5sum of servers ring and compare to local copy."`
	Quarantined bool   `short:"q" help:"Get cluster quarantine stats."`
	Replication bool   `short:"r" help:"Get replication stats."`
	Sharding    bool   `short:"s" help:"Check container sharding stats."`
	Unmounted   bool   `short:"u" help:"Check cluster for unmounted devices."`
	Updater     bool   `help:"Get updater stats."`
}

func main() {
	kong.Parse(&cli)
	switch {
	case cli.Diskusage && cli.Verbose:
		os.Stdout.Write(diskUsageVerboseData)
	case cli.Driveaudit && cli.Verbose:
		os.Stdout.Write(driveAuditVerboseData)
	case cli.MD5 && cli.Verbose:
		os.Stdout.Write(md5Data)
	case cli.Quarantined && cli.Verbose:
		os.Stdout.Write(quarantinedVerboseData)
	case cli.Replication && cli.Verbose:
		switch cli.ServerType {
		case "account":
			os.Stdout.Write(accountReplVerboseData)
		case "container":
			os.Stdout.Write(containerReplVerboseData)
		case "object":
			os.Stdout.Write(objectReplVerboseData)
		}
	case cli.Sharding:
		os.Stdout.Write(shardingVerboseData)
	case cli.Unmounted && cli.Verbose:
		os.Stdout.Write(unmountedVerboseData)
	case cli.Updater && cli.Verbose:
		switch cli.ServerType {
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
-> http://10.0.0.2:6000/recon/diskusage: [{u'device': u'sdb-15', u'avail': '', u'mounted': False, u'used': '', u'size': null}, {u'device': u'sdb-14', u'avail': 5608006787072, u'mounted': True, u'used': 391029243904, u'size': 5999036030976}, {u'device': u'sdb-03', u'avail': 5621251129344, u'mounted': True, u'used': 377784901632, u'size': 5999036030976}, {u'device': u'sdb-09', u'avail': 5646097563648, u'mounted': True, u'used': 352938467328, u'size': 5999036030976}, {u'device': u'sdb-10', u'avail': 5628910911488, u'mounted': True, u'used': 370125119488, u'size': 5999036030976}, {u'device': u'sdb-06', u'avail': 5615178706944, u'mounted': True, u'used': 383857324032, u'size': 5999036030976}, {u'device': u'sdb-07', u'avail': 5630346199040, u'mounted': True, u'used': 368689831936, u'size': 5999036030976}, {u'device': u'sdb-04', u'avail': 5624401117184, u'mounted': True, u'used': 374634913792, u'size': 5999036030976}, {u'device': u'sdb-05', u'avail': 5632579710976, u'mounted': True, u'used': 366456320000, u'size': 5999036030976}, {u'device': u'sdb-01', u'avail': 5639496806400, u'mounted': True, u'used': 359539224576, u'size': 5999036030976}, {u'device': u'sdb-08', u'avail': 5638742040576, u'mounted': True, u'used': 360293990400, u'size': 5999036030976}, {u'device': u'sdb-12', u'avail': 5635459399680, u'mounted': True, u'used': 363576631296, u'size': 5999036030976}, {u'device': u'sdb-11', u'avail': 5640890974208, u'mounted': True, u'used': 358145056768, u'size': 5999036030976}, {u'device': u'sdb-13', u'avail': 5622460055552, u'mounted': True, u'used': 376575975424, u'size': 5999036030976}, {u'device': u'sdb-02', u'avail': 5599360884736, u'mounted': True, u'used': 399675146240, u'size': 5999036030976}]
-> http://10.0.0.1:6000/recon/diskusage: [{u'device': u'sdb-14', u'avail': 5644861136896, u'mounted': True, u'used': 354174894080, u'size': 5999036030976}, {u'device': u'sdb-03', u'avail': 5635428171776, u'mounted': True, u'used': 363607859200, u'size': 5999036030976}, {u'device': u'sdb-09', u'avail': 5613408104448, u'mounted': True, u'used': 385627926528, u'size': 5999036030976}, {u'device': u'sdb-10', u'avail': 5626034053120, u'mounted': True, u'used': 373001977856, u'size': 5999036030976}, {u'device': u'sdb-06', u'avail': 5627536703488, u'mounted': True, u'used': 371499327488, u'size': 5999036030976}, {u'device': u'sdb-07', u'avail': 5612064350208, u'mounted': True, u'used': 386971680768, u'size': 5999036030976}, {u'device': u'sdb-04', u'avail': 5626924462080, u'mounted': True, u'used': 372111568896, u'size': 5999036030976}, {u'device': u'sdb-05', u'avail': 5619981594624, u'mounted': True, u'used': 379054436352, u'size': 5999036030976}, {u'device': u'sdb-01', u'avail': 5619673370624, u'mounted': True, u'used': 379362660352, u'size': 5999036030976}, {u'device': u'sdb-08', u'avail': 5639914979328, u'mounted': True, u'used': 359121051648, u'size': 5999036030976}, {u'device': u'sdb-12', u'avail': 5615940825088, u'mounted': True, u'used': 383095205888, u'size': 5999036030976}, {u'device': u'sdb-11', u'avail': 5628679016448, u'mounted': True, u'used': 370357014528, u'size': 5999036030976}, {u'device': u'sdb-13', u'avail': 5631554600960, u'mounted': True, u'used': 367481430016, u'size': 5999036030976}, {u'device': u'sdb-02', u'avail': 5626821857280, u'mounted': True, u'used': 372214173696, u'size': 5999036030976}]
Distribution Graph:
  5%    5 ***************
  6%   23 *********************************************************************
Disk usage: space used: 10421003354112 of 167973008867328
Disk usage: space free: 157552005513216 of 167973008867328
Disk usage: lowest: 5.88%, highest: 6.66%, avg: 6.20397492691%
===============================================================================`)

var md5Data = []byte(`===============================================================================
--> Starting reconnaissance on 4 hosts (object)
===============================================================================
[2021-05-04 17:57:31] Checking ring md5sums
-> On disk object.ring.gz md5sum: 12345
-> http://10.0.0.1:6000/recon/ringmd5: {'/path/to/account.ring.gz': '12345', '/path/to/container.ring.gz': '12345', '/path/to/object.ring.gz': '12345'}
-> http://10.0.0.2:6000/recon/ringmd5: {'/path/to/account.ring.gz': '12345', '/path/to/container.ring.gz': '12345', '/path/to/object.ring.gz': '12345'}
-> http://10.0.0.3:6000/recon/ringmd5: {'/path/to/account.ring.gz': '12345', '/path/to/container.ring.gz': '12345', '/path/to/object.ring.gz': '12345'}
-> http://10.0.0.4:6000/recon/ringmd5: {'/path/to/account.ring.gz': '12345', '/path/to/container.ring.gz': '12345', '/path/to/object.ring.gz': '12345'}
-> http://10.0.0.1:6000/recon/ringmd5 matches.
-> http://10.0.0.2:6000/recon/ringmd5 matches.
-> http://10.0.0.3:6000/recon/ringmd5 matches.
-> http://10.0.0.4:6000/recon/ringmd5 matches.
4/4 hosts matched, 0 error[s] while checking hosts.
===============================================================================
[2021-05-04 17:57:31] Checking swift.conf md5sum
-> On disk swift.conf md5sum: 12345
-> http://10.0.0.1:6000/recon/swiftconfmd5: {'/path/to/swift.conf': '12345'}
-> http://10.0.0.2:6000/recon/swiftconfmd5: {'/path/to/swift.conf': '12345'}
-> http://10.0.0.3:6000/recon/swiftconfmd5: {'/path/to/swift.conf': '12345'}
-> http://10.0.0.4:6000/recon/swiftconfmd5: {'/path/to/swift.conf': '12345'}
-> http://10.0.0.1:6000/recon/swiftconfmd5 matches.
-> http://10.0.0.2:6000/recon/swiftconfmd5 matches.
-> http://10.0.0.3:6000/recon/swiftconfmd5 matches.
-> http://10.0.0.4:6000/recon/swiftconfmd5 matches.
4/4 hosts matched, 0 error[s] while checking hosts.
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
-> http://10.0.0.3:6002/recon/replication/account: {'replication_time': None, 'replication_stats': None, 'replication_last': None}
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
-> http://10.0.0.3:6002/recon/replication/container: {'replication_time': None, 'replication_stats': None, 'replication_last': None}
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
-> http://10.0.0.3:6002/recon/replication/object: {'replication_time': None, 'replication_stats': None, 'replication_last': None}
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

var shardingVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 8 hosts (container)
===============================================================================
[2023-10-19 05:56:07] Checking on sharders
-> http://10.0.0.1:6001/recon/sharding: {'sharding_stats': {'attempted': 0, 'deferred': 0, 'diff': 0, 'diff_capped': 0, 'empty': 0, 'failure': 0, 'hashmatch': 0, 'no_change': 0, 'remote_merge': 0, 'remove': 0, 'rsync': 0, 'sharding': {'audit_root': {'attempted': 0, 'failure': 0, 'has_overlap': 0, 'num_overlap': 0, 'success': 0}, 'audit_shard': {'attempted': 12, 'failure': 0, 'success': 12}, 'cleaved': {'attempted': 0, 'failure': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'created': {'attempted': 0, 'failure': 0, 'success': 0}, 'misplaced': {'attempted': 12, 'failure': 0, 'found': 0, 'placed': 0, 'success': 12, 'unplaced': 0}, 'scanned': {'attempted': 0, 'failure': 0, 'found': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'sharding_candidates': {'found': 7, 'top': [{'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 1761173504, 'meta_timestamp': '1697694896.75479', 'node_index': 2, 'object_count': 10000000, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 882626560, 'meta_timestamp': '1697694910.47519', 'node_index': 1, 'object_count': 5000000, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 608010240, 'meta_timestamp': '1697694917.91514', 'node_index': 1, 'object_count': 2828027, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/versionssharding3'}, {'account': 'AUTH_ACCOUNT', 'container': 'sharding3', 'file_size': 1028579328, 'meta_timestamp': '1697694905.05529', 'node_index': 0, 'object_count': 2818948, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/sharding3'}, {'account': '.shards_AUTH_ACCOUNT', 'container': 'sharding-1695830239.64849-2', 'file_size': 251801600, 'meta_timestamp': '1697694923.15504', 'node_index': 1, 'object_count': 1448062, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/sharding'}]}, 'shrinking_candidates': {'found': 0, 'top': []}, 'visited': {'attempted': 12, 'completed': 0, 'failure': 0, 'skipped': 2983, 'success': 12}}, 'start': 1697694877.627436, 'success': 0, 'ts_repl': 0}, 'sharding_time': 59.88784885406494, 'sharding_last': 1697694937.5152848}
-> http://10.0.0.2:6001/recon/sharding: {'sharding_stats': {'attempted': 0, 'deferred': 0, 'diff': 0, 'diff_capped': 0, 'empty': 0, 'failure': 0, 'hashmatch': 0, 'no_change': 0, 'remote_merge': 0, 'remove': 0, 'rsync': 0, 'sharding': {'audit_root': {'attempted': 1, 'failure': 0, 'has_overlap': 0, 'num_overlap': 0, 'success': 1}, 'audit_shard': {'attempted': 6, 'failure': 0, 'success': 6}, 'cleaved': {'attempted': 0, 'failure': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'created': {'attempted': 0, 'failure': 0, 'success': 0}, 'misplaced': {'attempted': 7, 'failure': 0, 'found': 0, 'placed': 0, 'success': 7, 'unplaced': 0}, 'scanned': {'attempted': 0, 'failure': 0, 'found': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'sharding_candidates': {'found': 5, 'top': [{'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 2607591424, 'meta_timestamp': '1697694886.69780', 'node_index': 1, 'object_count': 15000000, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 1028476928, 'meta_timestamp': '1697694906.27765', 'node_index': 2, 'object_count': 2818948, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 343764992, 'meta_timestamp': '1697694910.25772', 'node_index': 0, 'object_count': 1983069, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256'}, {'account': 'AUTH_ACCOUNT', 'container': 'repo', 'file_size': 510939136, 'meta_timestamp': '1697694862.67756', 'node_index': 2, 'object_count': 1088685, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/repo'}, {'account': 'AUTH_ACCOUNT', 'container': 'atd-container1', 'file_size': 176001024, 'meta_timestamp': '1697694890.77763', 'node_index': 0, 'object_count': 1007650, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/atd-container1'}]}, 'shrinking_candidates': {'found': 0, 'top': []}, 'visited': {'attempted': 7, 'completed': 0, 'failure': 0, 'skipped': 2915, 'success': 7}}, 'start': 1697694856.1463997, 'success': 0, 'ts_repl': 0}, 'sharding_time': 58.43189835548401, 'sharding_last': 1697694914.578298}
-> http://10.0.0.3:6001/recon/sharding: {'sharding_stats': {'attempted': 0, 'deferred': 0, 'diff': 0, 'diff_capped': 0, 'empty': 0, 'failure': 0, 'hashmatch': 0, 'no_change': 0, 'remote_merge': 0, 'remove': 0, 'rsync': 0, 'sharding': {'audit_root': {'attempted': 1, 'failure': 0, 'has_overlap': 0, 'num_overlap': 0, 'success': 1}, 'audit_shard': {'attempted': 11, 'failure': 0, 'success': 11}, 'cleaved': {'attempted': 0, 'failure': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'created': {'attempted': 0, 'failure': 0, 'success': 0}, 'misplaced': {'attempted': 12, 'failure': 0, 'found': 0, 'placed': 0, 'success': 12, 'unplaced': 0}, 'scanned': {'attempted': 0, 'failure': 0, 'found': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'sharding_candidates': {'found': 3, 'top': [{'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 2030092288, 'meta_timestamp': '1697694865.43349', 'node_index': 1, 'object_count': 11473593, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 608059392, 'meta_timestamp': '1697694894.67352', 'node_index': 0, 'object_count': 2828027, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 175931392, 'meta_timestamp': '1697694921.21356', 'node_index': 1, 'object_count': 1007650, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/atd-container1'}]}, 'shrinking_candidates': {'found': 0, 'top': []}, 'visited': {'attempted': 12, 'completed': 0, 'failure': 0, 'skipped': 2942, 'success': 12}}, 'start': 1697694862.7463152, 'success': 0, 'ts_repl': 0}, 'sharding_time': 59.06757164001465, 'sharding_last': 1697694921.813887}
-> http://10.0.0.4:6001/recon/sharding: {'sharding_stats': {'attempted': 0, 'deferred': 0, 'diff': 0, 'diff_capped': 0, 'empty': 0, 'failure': 0, 'hashmatch': 0, 'no_change': 0, 'remote_merge': 0, 'remove': 0, 'rsync': 0, 'sharding': {'audit_root': {'attempted': 2, 'failure': 0, 'has_overlap': 0, 'num_overlap': 0, 'success': 2}, 'audit_shard': {'attempted': 8, 'failure': 0, 'success': 8}, 'cleaved': {'attempted': 0, 'failure': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'created': {'attempted': 0, 'failure': 0, 'success': 0}, 'misplaced': {'attempted': 10, 'failure': 0, 'found': 0, 'placed': 0, 'success': 10, 'unplaced': 0}, 'scanned': {'attempted': 0, 'failure': 0, 'found': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'sharding_candidates': {'found': 6, 'top': [{'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 2607591424, 'meta_timestamp': '1697713537.19859', 'node_index': 1, 'object_count': 15000000, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 1028476928, 'meta_timestamp': '1697713547.42125', 'node_index': 2, 'object_count': 2818948, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': '.shards_AUTH_ACCOUNT', 'container': 'container-sharding-1697704617.79400-0', 'file_size': 259579904, 'meta_timestamp': '1697713503.02092', 'node_index': 2, 'object_count': 1496509, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256'}, {'account': '.shards_AUTH_ACCOUNT', 'container': '2nd-sharding-256-fe05da69b6bbe90618870d22db0a9d53-1697704617.79400-1', 'file_size': 256335872, 'meta_timestamp': '1697713525.04081', 'node_index': 2, 'object_count': 1477936, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256'}, {'account': 'AUTH_ACCOUNT', 'container': 'repo', 'file_size': 510939136, 'meta_timestamp': '1697713504.29706', 'node_index': 2, 'object_count': 1088685, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/repo'}]}, 'sharding_in_progress': {'all': [{'account': 'AUTH_ACCOUNT', 'active': 2, 'cleaved': 0, 'container': '2nd-sharding-256', 'created': 0, 'db_state': 'sharded', 'error': 'None', 'file_size': 53248, 'found': 0, 'meta_timestamp': '1697713521.88929', 'node_index': 0, 'object_count': 2974445, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256', 'state': 'sharded'}]}, 'shrinking_candidates': {'found': 0, 'top': []}, 'visited': {'attempted': 10, 'completed': 0, 'failure': 0, 'skipped': 2922, 'success': 10}}, 'start': 1697713502.8096673, 'success': 0, 'ts_repl': 0}, 'sharding_time': 58.6283392906189, 'sharding_last': 1697713561.4380066}
-> http://10.0.0.5:6001/recon/sharding: {'sharding_stats': {'attempted': 0, 'deferred': 0, 'diff': 0, 'diff_capped': 0, 'empty': 0, 'failure': 0, 'hashmatch': 0, 'no_change': 0, 'remote_merge': 0, 'remove': 0, 'rsync': 0, 'sharding': {'audit_root': {'attempted': 1, 'failure': 0, 'has_overlap': 0, 'num_overlap': 0, 'success': 1}, 'audit_shard': {'attempted': 11, 'failure': 0, 'success': 11}, 'cleaved': {'attempted': 0, 'failure': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'created': {'attempted': 0, 'failure': 0, 'success': 0}, 'misplaced': {'attempted': 13, 'failure': 0, 'found': 0, 'placed': 0, 'success': 13, 'unplaced': 0}, 'scanned': {'attempted': 0, 'failure': 0, 'found': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'sharding_candidates': {'found': 9, 'top': [{'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 2607628288, 'meta_timestamp': '1697713530.46491', 'node_index': 0, 'object_count': 15000000, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': '\x00versions\x00sharding', 'file_size': 608043008, 'meta_timestamp': '1697713499.62458', 'node_index': 2, 'object_count': 2828027, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/\x00versions\x00sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 1028464640, 'meta_timestamp': '1697713529.24476', 'node_index': 1, 'object_count': 2818948, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': '.shards_AUTH_ACCOUNT', 'container': 'sharding-c4f60a871b95e04ff96bb2cbbf1f96e6-1695830239.64849-3', 'file_size': 292560896, 'meta_timestamp': '1697713551.72490', 'node_index': 1, 'object_count': 1681328, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/sharding'}, {'account': '.shards_AUTH_ACCOUNT', 'container': '2nd-sharding-256-1', 'file_size': 256315392, 'meta_timestamp': '1697713548.62538', 'node_index': 0, 'object_count': 1477936, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256'}]}, 'shrinking_candidates': {'found': 0, 'top': []}, 'visited': {'attempted': 13, 'completed': 0, 'failure': 0, 'skipped': 3440, 'success': 13}}, 'start': 1697713487.3573852, 'success': 0, 'ts_repl': 0}, 'sharding_time': 69.04765748977661, 'sharding_last': 1697713556.4050426}
-> http://10.0.0.6:6001/recon/sharding: {'sharding_stats': {'attempted': 0, 'deferred': 0, 'diff': 0, 'diff_capped': 0, 'empty': 0, 'failure': 0, 'hashmatch': 0, 'no_change': 0, 'remote_merge': 0, 'remove': 0, 'rsync': 0, 'sharding': {'audit_root': {'attempted': 2, 'failure': 0, 'has_overlap': 0, 'num_overlap': 0, 'success': 2}, 'audit_shard': {'attempted': 8, 'failure': 0, 'success': 8}, 'cleaved': {'attempted': 0, 'failure': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'created': {'attempted': 0, 'failure': 0, 'success': 0}, 'misplaced': {'attempted': 10, 'failure': 0, 'found': 0, 'placed': 0, 'success': 10, 'unplaced': 0}, 'scanned': {'attempted': 0, 'failure': 0, 'found': 0, 'max_time': 0, 'min_time': 0, 'success': 0}, 'sharding_candidates': {'found': 6, 'top': [{'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 2607591424, 'meta_timestamp': '1697713537.19859', 'node_index': 1, 'object_count': 15000000, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': 'AUTH_ACCOUNT', 'container': 'container-sharding', 'file_size': 1028476928, 'meta_timestamp': '1697713547.42125', 'node_index': 2, 'object_count': 2818948, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/container-sharding'}, {'account': '.shards_AUTH_ACCOUNT', 'container': 'container-sharding-1697704617.79400-0', 'file_size': 259579904, 'meta_timestamp': '1697713503.02092', 'node_index': 2, 'object_count': 1496509, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256'}, {'account': '.shards_AUTH_ACCOUNT', 'container': '2nd-sharding-256-fe05da69b6bbe90618870d22db0a9d53-1697704617.79400-1', 'file_size': 256335872, 'meta_timestamp': '1697713525.04081', 'node_index': 2, 'object_count': 1477936, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256'}, {'account': 'AUTH_ACCOUNT', 'container': 'repo', 'file_size': 510939136, 'meta_timestamp': '1697713504.29706', 'node_index': 2, 'object_count': 1088685, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/repo'}]}, 'sharding_in_progress': {'all': [{'account': 'AUTH_ACCOUNT', 'active': 2, 'cleaved': 0, 'container': '2nd-sharding-256', 'created': 0, 'db_state': 'sharded', 'error': 'something other than 0', 'file_size': 53248, 'found': 0, 'meta_timestamp': '1697713521.88929', 'node_index': 0, 'object_count': 2974445, 'path': '/path/to/db.db', 'root': 'AUTH_ACCOUNT/2nd-sharding-256', 'state': 'sharded'}]}, 'shrinking_candidates': {'found': 0, 'top': []}, 'visited': {'attempted': 10, 'completed': 0, 'failure': 0, 'skipped': 2922, 'success': 10}}, 'start': 1697713502.8096673, 'success': 0, 'ts_repl': 0}, 'sharding_time': 58.6283392906189, 'sharding_last': 1697713561.4380066}
[sharding_time] low: 57, high: 69, avg: 61.5, total: 491, Failed: 0.0%, no_result: 0, reported: 8
[attempted] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 8
[failure] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 8
[success] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 8
Oldest completion was 2023-10-19 05:55:04 (1 minutes ago) by 10.246.204.66:6001.
Most recent completion was 2023-10-19 05:56:07 (0 seconds ago) by 10.245.58.91:6001.
===============================================================================`)
