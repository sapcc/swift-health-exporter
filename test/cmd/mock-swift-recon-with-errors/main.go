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
[2020-01-14 12:55:00] Checking disk usage now
-> http://10.0.0.1:6000/recon/diskusage: [{u'device': u'sdb-14', u'avail': 5646060654592, u'mounted': True, u'used': 352975376384, u'size': 5999036030976}, {u'device': u'sdb-03', u'avail': 5634014085120, u'mounted': True, u'used': 365021945856, u'size': 5999036030976}, {u'device': u'sdb-09', u'avail': 5611860353024, u'mounted': True, u'used': 387175677952, u'size': 5999036030976}, {u'device': u'sdb-10', u'avail': 5629413326848, u'mounted': True, u'used': 369622704128, u'size': 5999036030976}, {u'device': u'sdb-06', u'avail': 5628006096896, u'mounted': True, u'used': 371029934080, u'size': 5999036030976}, {u'device': u'sdb-07', u'avail': 5612116451328, u'mounted': True, u'used': 386919579648, u'size': 5999036030976}, {u'device': u'sdb-04', u'avail': 5624243228672, u'mounted': True, u'used': 374792802304, u'size': 5999036030976}, {u'device': u'sdb-05', u'avail': 5616945328128, u'mounted': True, u'used': 382090702848, u'size': 5999036030976}, {u'device': u'sdb-01', u'avail': 5621843275776, u'mounted': True, u'used': 377192755200, u'size': 5999036030976}, {u'device': u'sdb-08', u'avail': 5636845748224, u'mounted': True, u'used': 362190282752, u'size': 5999036030976}, {u'device': u'sdb-12', u'avail': 5611945791488, u'mounted': True, u'used': 387090239488, u'size': 5999036030976}, {u'device': u'sdb-11', u'avail': 5627626954752, u'mounted': True, u'used': 371409076224, u'size': 5999036030976}, {u'device': u'sdb-13', u'avail': 5632584495104, u'mounted': True, u'used': 366451535872, u'size': 5999036030976}, {u'device': u'sdb-02', u'avail': 5632245743616, u'mounted': True, u'used': 366790287360, u'size': 5999036030976}]
-> http://10.0.0.2:6000/recon/diskusage: <urlopen error [Errno 111] ECONNREFUSED>
Distribution Graph:
  5%    1 *****
  6%   13 *********************************************************************
Disk usage: space used: 5220752900096 of 83986504433664
Disk usage: space free: 78765751533568 of 83986504433664
Disk usage: lowest: 5.88%, highest: 6.45%, avg: 6.21618072487%
===============================================================================`)

var md5Data = []byte(`===============================================================================
--> Starting reconnaissance on 4 hosts (object)
===============================================================================
[2021-05-04 17:57:31] Checking ring md5sums
-> On disk object.ring.gz md5sum: 12345
-> http://10.0.0.2:6000/recon/ringmd5: <urlopen error timed out>
-> http://10.0.0.1:6000/recon/ringmd5: {'/path/to/account.ring.gz': '12345', '/path/to/container.ring.gz': '12345', '/path/to/object.ring.gz': '12345'}
-> http://10.0.0.3:6000/recon/ringmd5: {'/path/to/account.ring.gz': '12345', '/path/to/container.ring.gz': '12345', '/path/to/object.ring.gz': '12345'}
-> http://10.0.0.1:6000/recon/ringmd5 matches.
-> http://10.0.0.3:6000/recon/ringmd5 matches.
!! http://10.0.0.4:6000/recon/ringmd5 (/path/to/account.ring.gz => 54321) doesn't match on disk md5sum
!! http://10.0.0.4:6000/recon/ringmd5 (/path/to/container.ring.gz => 54321) doesn't match on disk md5sum
!! http://10.0.0.4:6000/recon/ringmd5 (/path/to/object.ring.gz => 54321) doesn't match on disk md5sum
2/4 hosts matched, 1 error[s] while checking hosts.
===============================================================================
[2021-05-04 17:57:31] Checking swift.conf md5sum
-> On disk swift.conf md5sum: 12345
-> http://10.0.0.2:6000/recon/ringmd5: <urlopen error [Errno 111] ECONNREFUSED>
-> http://10.0.0.1:6000/recon/swiftconfmd5: {'/path/to/swift.conf': '12345'}
-> http://10.0.0.3:6000/recon/swiftconfmd5: {'/path/to/swift.conf': '12345'}
-> http://10.0.0.1:6000/recon/swiftconfmd5 matches.
-> http://10.0.0.3:6000/recon/swiftconfmd5 matches.
-> http://10.0.0.4:6000/recon/swiftconfmd5: (/path/to/swift.conf => 54321) doesn't match on disk md5sum
2/4 hosts matched, 1 error[s] while checking hosts.
===============================================================================`)

var containerUpdaterVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (container)
===============================================================================
[2020-01-14 13:08:17] Checking updater times
-> http://10.0.0.1:6001/recon/updater/container: {u'container_updater_sweep': 54.06525897979736}
-> http://10.0.0.2:6001/recon/updater/container: <urlopen error timed out>
[updater_last_sweep] low: 54, high: 54, avg: 54.1, total: 54, Failed: 0.0%, no_result: 0, reported: 1
===============================================================================`)

var objectUpdaterVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2020-01-14 13:08:07] Checking updater times
-> http://10.0.0.1:6000/recon/updater/object: {u'object_updater_sweep': 1.863548994064331}
-> http://10.0.0.2:6000/recon/updater/object: <urlopen error timed out>
[updater_last_sweep] low: 1, high: 1, avg: 1.9, total: 1, Failed: 0.0%, no_result: 0, reported: 1
===============================================================================`)

var accountReplVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (account)
===============================================================================
[2020-01-14 13:07:32] Checking on replication
-> http://10.0.0.1:6002/recon/replication/account: {u'replication_last': 1579007237.099724, u'replication_stats': {u'no_change': 408, u'rsync': 0, u'success': 410, u'failure': 816, u'attempted': 613, u'ts_repl': 0, u'remove': 0, u'remote_merge': 0, u'diff_capped': 0, u'deferred': 0, u'hashmatch': 0, u'failure_nodes': {u'10.0.0.2': {u'sdb-11': 73, u'sdb-10': 57, u'sdb-13': 53, u'sdb-12': 49, u'sdb-14': 48, u'sdb-08': 54, u'sdb-09': 56, u'sdb-06': 64, u'sdb-07': 60, u'sdb-04': 67, u'sdb-05': 58, u'sdb-02': 63, u'sdb-03': 59, u'sdb-01': 55}}, u'diff': 2, u'start': 1579007213.676877, u'empty': 0}, u'replication_time': 23.422847032546997}
-> http://10.0.0.2:6002/recon/replication/account: <urlopen error timed out>
[replication_failure] low: 816, high: 816, avg: 816.0, total: 816, Failed: 0.0%, no_result: 0, reported: 1
[replication_success] low: 410, high: 410, avg: 410.0, total: 410, Failed: 0.0%, no_result: 0, reported: 1
[replication_time] low: 23, high: 23, avg: 23.4, total: 23, Failed: 0.0%, no_result: 0, reported: 1
[replication_attempted] low: 613, high: 613, avg: 613.0, total: 613, Failed: 0.0%, no_result: 0, reported: 1
Oldest completion was 2020-01-14 13:07:17 (20 seconds ago) by 10.0.0.1:6002.
Most recent completion was 2020-01-14 13:07:17 (20 seconds ago) by 10.0.0.1:6002.
===============================================================================`)

var containerReplVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (container)
===============================================================================
[2020-01-14 13:07:42] Checking on replication
-> http://10.0.0.1:6001/recon/replication/container: {u'replication_last': 1579007236.617117, u'replication_stats': {u'no_change': 7963, u'rsync': 0, u'success': 7966, u'failure': 814, u'attempted': 4390, u'ts_repl': 0, u'remove': 0, u'remote_merge': 0, u'diff_capped': 0, u'deferred': 0, u'hashmatch': 0, u'failure_nodes': {u'10.0.0.2': {u'sdb-11': 42, u'sdb-10': 76, u'sdb-13': 69, u'sdb-12': 47, u'sdb-14': 73, u'sdb-08': 63, u'sdb-09': 41, u'sdb-06': 51, u'sdb-07': 55, u'sdb-04': 61, u'sdb-05': 66, u'sdb-02': 61, u'sdb-03': 71, u'sdb-01': 38}}, u'diff': 3, u'start': 1579007138.241347, u'empty': 0}, u'replication_time': 98.37576985359192}
-> http://10.0.0.2:6001/recon/replication/container: <urlopen error timed out>
[replication_failure] low: 814, high: 814, avg: 814.0, total: 814, Failed: 0.0%, no_result: 0, reported: 1
[replication_success] low: 7966, high: 7966, avg: 7966.0, total: 7966, Failed: 0.0%, no_result: 0, reported: 1
[replication_time] low: 98, high: 98, avg: 98.4, total: 98, Failed: 0.0%, no_result: 0, reported: 1
[replication_attempted] low: 4390, high: 4390, avg: 4390.0, total: 4390, Failed: 0.0%, no_result: 0, reported: 1
Oldest completion was 2020-01-14 13:07:16 (30 seconds ago) by 10.0.0.1:6001.
Most recent completion was 2020-01-14 13:07:16 (30 seconds ago) by 10.0.0.1:6001.
===============================================================================`)

var objectReplVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2020-01-14 13:07:56] Checking on replication
-> http://10.0.0.1:6000/recon/replication/object: {u'replication_last': 1579006461.81673, u'replication_stats': {u'rsync': 9, u'success': 168394, u'failure': 28214, u'attempted': 98304, u'remove': 0, u'suffix_count': 1267858, u'start': 1535617859.516976, u'hashmatch': 168393, u'failure_nodes': {u'10.0.0.1': {u'sdb-12': 3, u'sdb-02': 1}, u'10.0.0.2': {u'sdb-11': 2013, u'sdb-10': 2021, u'sdb-13': 2037, u'sdb-12': 2061, u'sdb-14': 1976, u'sdb-08': 1915, u'sdb-09': 2018, u'sdb-06': 2016, u'sdb-07': 2008, u'sdb-04': 2112, u'sdb-05': 2007, u'sdb-02': 1988, u'sdb-03': 2041, u'sdb-01': 1997}}, u'suffix_sync': 1, u'suffix_hash': 7}, u'replication_time': 5.449508202075958, u'object_replication_last': 1579006461.81673, u'object_replication_time': 5.449508202075958}
-> http://10.0.0.2:6000/recon/replication/object: <urlopen error timed out>
[replication_failure] low: 28214, high: 28214, avg: 28214.0, total: 28214, Failed: 0.0%, no_result: 0, reported: 1
[replication_success] low: 168394, high: 168394, avg: 168394.0, total: 168394, Failed: 0.0%, no_result: 0, reported: 1
[replication_time] low: 5, high: 5, avg: 5.4, total: 5, Failed: 0.0%, no_result: 0, reported: 1
[replication_attempted] low: 98304, high: 98304, avg: 98304.0, total: 98304, Failed: 0.0%, no_result: 0, reported: 1
Oldest completion was 2020-01-14 12:54:21 (13 minutes ago) by 10.0.0.1:6000.
Most recent completion was 2020-01-14 12:54:21 (13 minutes ago) by 10.0.0.1:6000.
===============================================================================`)

var quarantinedVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2020-01-14 12:56:01] Checking quarantine
-> http://10.0.0.1:6000/recon/quarantined: {u'objects': 0, u'accounts': 0, u'containers': 0, u'policies': {}}
-> http://10.0.0.2:6000/recon/quarantined: <urlopen error timed out>
[quarantined_objects] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 1
[quarantined_accounts] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 1
[quarantined_containers] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 1
===============================================================================`)

var unmountedVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2020-01-14 12:57:02] Getting unmounted drives from 2 hosts...
-> http://10.0.0.1:6000/recon/unmounted: []
-> http://10.0.0.2:6000/recon/unmounted: <urlopen error timed out>
===============================================================================`)

var driveAuditVerboseData = []byte(`===============================================================================
--> Starting reconnaissance on 2 hosts (object)
===============================================================================
[2020-01-14 12:55:44] Checking drive-audit errors
-> http://10.0.0.1:6000/recon/driveaudit: {u'drive_audit_errors': 0}
-> http://10.0.0.2:6000/recon/driveaudit: <urlopen error timed out>
[drive_audit_errors] low: 0, high: 0, avg: 0.0, total: 0, Failed: 0.0%, no_result: 0, reported: 1
===============================================================================`)
