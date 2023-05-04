package util

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/process"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
	"gopkg.in/mgo.v2/bson"
)

const (
	TypeIsDBGrid    = "isdbgrid"
	TypeMongos      = "mongos"
	TypeMongod      = "mongod"
	TypeShardServer = "shardsvr"

	milliToSeconds              = 1000
	shardingNotEnabledErrorCode = 203
	ErrNotYetInitialized        = int32(94)
	ErrNoReplicationEnabled     = int32(76)
	ErrNotPrimaryOrSecondary    = int32(13436)
)

var (
	CannotGetQueryError     = errors.New("cannot get query field from the profile document (it is not a map)")
	ShardingNotEnabledError = errors.New("sharding not enabled")

	ErrCannotGetProcess   = fmt.Errorf("cannot get process")
	ErrCannotGetClusterID = fmt.Errorf("cannot get cluster ID")
)

func GetReplicasetMembers(ctx context.Context, clientOptions *options.ClientOptions) ([]proto.Members, error) {
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get a new client for GetReplicasetMembers")
	}

	if err := client.Connect(ctx); err != nil {
		return nil, errors.Wrap(err, "cannot connect to MongoDB")
	}

	hostnames, err := GetHostnames(ctx, client)
	if err != nil {
		return nil, err
	}

	if err := client.Disconnect(ctx); err != nil {
		return nil, errors.Wrapf(err, "cannot disconnect from %v", clientOptions.Hosts)
	}

	membersMap := make(map[string]proto.Members)
	members := []proto.Members{}

	for _, hostname := range hostnames {
		client, err = GetClientForHost(clientOptions, hostname)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot get a new client to connect to %s", hostname)
		}

		if err := client.Connect(ctx); err != nil {
			return nil, errors.Wrapf(err, "cannot connect to %s", hostname)
		}

		cmdOpts := proto.CommandLineOptions{}
		// Not always we can get this info. For examples, we cannot get this for hidden hosts so
		// if there is an error, just ignore it
		res := client.Database("admin").RunCommand(ctx, primitive.D{
			{Key: "getCmdLineOpts", Value: 1},
			{Key: "recordStats", Value: 1},
		})
		if res.Err() == nil {
			if err := res.Decode(&cmdOpts); err != nil {
				return nil, errors.Wrapf(err, "cannot decode getCmdLineOpts response for host %s", hostname)
			}
		}

		rss := proto.ReplicaSetStatus{}

		res = client.Database("admin").RunCommand(ctx, primitive.M{"replSetGetStatus": 1})
		if res.Err() != nil {
			m := proto.Members{
				Name: hostname,
			}
			m.StateStr = strings.ToUpper(cmdOpts.Parsed.Sharding.ClusterRole)

			if serverStatus, err := GetServerStatus(ctx, client); err == nil {
				m.ID = serverStatus.Pid
				m.StorageEngine = serverStatus.StorageEngine
			}

			membersMap[m.Name] = m

			continue // If a host is a mongos we cannot get info but is not a real error
		}

		if err := res.Decode(&rss); err != nil {
			return nil, errors.Wrap(err, "cannot decode replSetGetStatus response")
		}

		for _, m := range rss.Members {
			if _, ok := membersMap[m.Name]; ok {
				continue // already exists
			}

			m.Set = rss.Set

			if serverStatus, err := GetServerStatus(ctx, client); err == nil {
				m.ID = serverStatus.Pid
				m.StorageEngine = serverStatus.StorageEngine

				if cmdOpts.Parsed.Sharding.ClusterRole != "" {
					m.StateStr = cmdOpts.Parsed.Sharding.ClusterRole + "/" + m.StateStr
				}

				m.StateStr = strings.ToUpper(m.StateStr)
			}

			membersMap[m.Name] = m
		}

		client.Disconnect(ctx) //nolint
	}

	for _, member := range membersMap {
		members = append(members, member)
	}

	sort.Slice(members, func(i, j int) bool { return members[i].Name < members[j].Name })

	return members, nil
}

func GetHostnames(ctx context.Context, client *mongo.Client) ([]string, error) {
	// Probably we are connected to an individual member of a replica set
	rss := proto.ReplicaSetStatus{}

	res := client.Database("admin").RunCommand(ctx, primitive.M{"replSetGetStatus": 1})
	if res.Err() == nil {
		if err := res.Decode(&rss); err != nil {
			return nil, errors.Wrap(err, "cannot decode replSetGetStatus response for GetHostnames")
		}

		return buildHostsListFromReplStatus(rss), nil
	}

	// Try getShardMap first. If we are connected to a mongos it will return
	// all hosts, including config hosts
	var shardsMap proto.ShardsMap

	smRes := client.Database("admin").RunCommand(ctx, primitive.M{"getShardMap": 1})
	if smRes.Err() != nil {
		if e, ok := smRes.Err().(mongo.CommandError); ok && e.Code == shardingNotEnabledErrorCode {
			return nil, ShardingNotEnabledError // standalone instance
		}

		return nil, errors.Wrap(smRes.Err(), "cannot getShardMap for GetHostnames")
	}

	if err := smRes.Decode(&shardsMap); err != nil {
		return nil, errors.Wrap(err, "cannot decode getShardMap result for GetHostnames")
	}

	if len(shardsMap.Map) > 0 {
		// if the only element getShardMap returns is the list of config servers,
		// it means we are connected to a replicaSet member and getShardMap is not
		// the answer we want.
		if _, ok := shardsMap.Map["config"]; ok {
			return buildHostsListFromShardMap(shardsMap), nil
		}
	}

	// Some MongoDB servers won't return ShardingNotEnabledError for stand alone instances.
	return nil, nil // standalone instance
}

/*
   "members" : [
            {
                    "_id" : 0,
                    "name" : "localhost:17001",
                    "health" : 1,
                    "state" : 1,
                    "stateStr" : "PRIMARY",
                    "uptime" : 4700,
                    "optime" : Timestamp(1486554836, 1),
                    "optimeDate" : ISODate("2017-02-08T11:53:56Z"),
                    "electionTime" : Timestamp(1486651810, 1),
                    "electionDate" : ISODate("2017-02-09T14:50:10Z"),
                    "configVersion" : 1,
                    "self" : true
            },
*/
func buildHostsListFromReplStatus(replStatus proto.ReplicaSetStatus) []string {
	hostnames := []string{}
	for _, member := range replStatus.Members {
		hostnames = append(hostnames, member.Name)
	}

	sort.Strings(hostnames) // to make testing easier

	return hostnames
}

/* Example
mongos> db.getSiblingDB('admin').runCommand('getShardMap')
{
  "map" : {
    "config" : "localhost:19001,localhost:19002,localhost:19003",
    "localhost:17001" : "r1/localhost:17001,localhost:17002,localhost:17003",
    "r1" : "r1/localhost:17001,localhost:17002,localhost:17003",
    "r1/localhost:17001,localhost:17002,localhost:17003" : "r1/localhost:17001,localhost:17002,localhost:17003",
  },
  "ok" : 1
}.
*/
func buildHostsListFromShardMap(shardsMap proto.ShardsMap) []string {
	hostnames := []string{}
	hm := make(map[string]bool)

	// Since shardMap can return repeated hosts in different rows, we need a Set
	// but there is no Set in Go so, we are going to create a map and the loop
	// through the keys to get a list of unique host names
	if shardsMap.Map != nil {
		for _, val := range shardsMap.Map {
			m := strings.Split(val, "/")
			hostsStr := ""

			switch len(m) {
			case 1:
				hostsStr = m[0] // there is no / in the hosts list
			case 2: //nolint
				hostsStr = m[1] // there is a / in the string. Remove the prefix until the / and keep the rest
			}
			// since there is no Sets in Go, build a map where the value is the map key
			hosts := strings.Split(hostsStr, ",")
			for _, host := range hosts {
				hm[host] = false
			}
		}

		for host := range hm {
			hostnames = append(hostnames, host)
		}
	}

	sort.Strings(hostnames)

	return hostnames
}

// GetShardedHosts is like GetHostnames but it uses listShards instead of getShardMap
// so it won't include config servers in the returned list.
func GetShardedHosts(ctx context.Context, client *mongo.Client) ([]string, error) {
	shardsInfo := &proto.ShardsInfo{}

	res := client.Database("admin").RunCommand(ctx, primitive.M{"listShards": 1})
	if res.Err() != nil {
		return nil, errors.Wrap(res.Err(), "cannot list shards")
	}

	if err := res.Decode(&shardsInfo); err != nil {
		return nil, errors.Wrap(err, "cannot decode listShards response")
	}

	hostnames := []string{}

	for _, shardInfo := range shardsInfo.Shards {
		m := strings.Split(shardInfo.Host, "/")
		h := strings.Split(m[1], ",")
		hostnames = append(hostnames, h[0])
	}

	return hostnames, nil
}

// GetServerStatus returns the server status by running serverStatus and recordStats
func GetServerStatus(ctx context.Context, client *mongo.Client) (proto.ServerStatus, error) {
	ss := proto.ServerStatus{}

	query := primitive.D{
		{Key: "serverStatus", Value: 1},
		{Key: "recordStats", Value: 1},
	}
	res := client.Database("admin").RunCommand(ctx, query)

	if res.Err() != nil {
		return ss, errors.Wrap(res.Err(), "GetHostInfo.serverStatus")
	}

	if err := res.Decode(&ss); err != nil {
		return ss, errors.Wrap(err, "cannot decode serverStatus/recordStats")
	}

	return ss, nil
}

func GetQueryField(doc proto.SystemProfile) (primitive.M, error) {
	// Proper way to detect if protocol used is "op_msg" or "op_command"
	// would be to look at "doc.Protocol" field,
	// however MongoDB 3.0 doesn't have that field
	// so we need to detect protocol by looking at actual data.
	query := doc.Query

	if len(doc.Command) > 0 { //nolint
		query = doc.Command

		if doc.Op == "update" || doc.Op == "remove" {
			if squery, ok := query.Map()["q"]; ok {
				// just an extra check to ensure this type assertion won't fail
				if ssquery, ok := squery.(primitive.M); ok {
					return ssquery, nil
				}

				return nil, CannotGetQueryError
			}
		}
	}

	// "query" in MongoDB 3.0 can look like this:
	// {
	//  	"op" : "query",
	//  	"ns" : "test.coll",
	//  	"query" : {
	//  		"a" : 1
	//  	},
	// 		...
	// }
	//
	// but also it can have "query" subkey like this:
	// {
	//  	"op" : "query",
	//  	"ns" : "test.coll",
	//  	"query" : {
	//  		"query" : {
	//  			"$and" : [
	//  			]
	//  		},
	//  		"orderby" : {
	//  			"k" : -1
	//  		}
	//  	},
	// 		...
	// }
	//
	if squery, ok := query.Map()["query"]; ok {
		// just an extra check to ensure this type assertion won't fail
		if ssquery, ok := squery.(primitive.M); ok {
			return ssquery, nil
		}

		return nil, CannotGetQueryError
	}

	// "query" in MongoDB 3.2+ is better structured and always has a "filter" subkey:
	if squery, ok := query.Map()["filter"]; ok {
		if ssquery, ok := squery.(primitive.M); ok {
			return ssquery, nil
		}

		return nil, CannotGetQueryError
	}

	// {"ns":"test.system.js","op":"query","query":{"find":"system.js"}}
	if len(query) == 1 && query[0].Key == "find" {
		return primitive.M{}, nil
	}

	return query.Map(), nil
}

// GetClientForHost returns a new *mongo.Client using a copy of the original connection options where
// the host is being replaced by the newHost and the connection is set to be direct to the instance.
func GetClientForHost(co *options.ClientOptions, newHost string) (*mongo.Client, error) {
	newOptions := options.MergeClientOptions(co, &options.ClientOptions{Hosts: []string{newHost}})
	newOptions.SetDirect(true)

	return mongo.NewClient(newOptions)
}

func GetHostInfo(ctx context.Context, client *mongo.Client) (*proto.GetHostInfo, error) {
	hi := proto.HostInfo{}
	if err := client.Database("admin").RunCommand(ctx, primitive.M{"hostInfo": 1}).Decode(&hi); err != nil {
		log.Printf("run('hostInfo') error: %s", err)
		return nil, errors.Wrap(err, "GetHostInfo.hostInfo")
	}

	cmdOpts := proto.CommandLineOptions{}
	query := primitive.D{{Key: "getCmdLineOpts", Value: 1}, {Key: "recordStats", Value: 1}}

	err := client.Database("admin").RunCommand(ctx, query).Decode(&cmdOpts)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get command line options")
	}

	ss := proto.ServerStatus{}
	query = primitive.D{{Key: "serverStatus", Value: 1}, {Key: "recordStats", Value: 1}}

	if err := client.Database("admin").RunCommand(ctx, query).Decode(&ss); err != nil {
		return nil, errors.Wrap(err, "GetHostInfo.serverStatus")
	}

	pi := proto.ProcInfo{}
	if err := fillProcInfo(int32(ss.Pid), &pi); err != nil {
		pi.Error = err
	}

	nodeType, _ := getNodeType(ctx, client)
	procCount, _ := countMongodProcesses()

	i := &proto.GetHostInfo{
		Hostname:          hi.System.Hostname,
		HostOsType:        hi.Os.Type,
		HostSystemCPUArch: hi.System.CpuArch,
		DBPath:            "", // Sets default. It will be overridden later if necessary

		ProcessName:      ss.Process,
		ProcProcessCount: procCount,
		Version:          ss.Version,
		NodeType:         nodeType,

		ProcPath:       pi.Path,
		ProcUserName:   pi.UserName,
		ProcCreateTime: pi.CreateTime,
	}
	if ss.Repl != nil {
		i.ReplicasetName = ss.Repl.SetName
	}

	if cmdOpts.Parsed.Storage.DbPath != "" {
		i.DBPath = cmdOpts.Parsed.Storage.DbPath
	}

	return i, nil
}

func ClusterID(ctx context.Context, client *mongo.Client) (string, error) {
	var cv proto.ConfigVersion
	if err := client.Database("config").Collection("version").FindOne(ctx, bson.M{}).Decode(&cv); err == nil {
		return cv.ClusterID.Hex(), nil
	}

	var si proto.ShardIdentity

	filter := bson.M{"_id": "shardIdentity"}

	if err := client.Database("admin").Collection("system.version").FindOne(ctx, filter).Decode(&si); err == nil {
		return si.ClusterID.Hex(), nil
	}

	rc, err := ReplicasetConfig(ctx, client)
	if err != nil {
		if e, ok := err.(mongo.CommandError); ok && IsReplicationNotEnabledError(e) {
			return "", nil
		}
		if _, ok := err.(topology.ServerSelectionError); ok {
			return "", nil
		}
		return "", err
	}

	return rc.Config.Settings.ReplicaSetID.Hex(), nil
}

func IsReplicationNotEnabledError(err mongo.CommandError) bool {
	return err.Code == ErrNotYetInitialized || err.Code == ErrNoReplicationEnabled ||
		err.Code == ErrNotPrimaryOrSecondary
}

func MyState(ctx context.Context, client *mongo.Client) (int, error) {
	var ms proto.MyState

	err := client.Database("admin").RunCommand(ctx, bson.M{"getDiagnosticData": 1}).Decode(&ms)
	if err != nil {
		return 0, err
	}

	return ms.Data.ReplicasetGetStatus.MyState, nil
}

func ReplicasetConfig(ctx context.Context, client *mongo.Client) (*proto.ReplicasetConfig, error) {
	var rs proto.ReplicasetConfig
	if err := client.Database("admin").RunCommand(ctx, bson.M{"replSetGetConfig": 1}).Decode(&rs); err != nil {
		return nil, err
	}

	return &rs, nil
}

func fillProcInfo(pid int32, procInfo *proto.ProcInfo) error {
	proc, err := process.NewProcess(pid)
	if err != nil {
		return errors.Wrapf(ErrCannotGetProcess, "%s", err)
	}

	ct, err := proc.CreateTime()
	if err != nil {
		return err
	}

	procInfo.CreateTime = time.Unix(ct/milliToSeconds, 0)

	if procInfo.Path, err = proc.Exe(); err != nil {
		return err
	}

	if procInfo.UserName, err = proc.Username(); err != nil {
		return err
	}

	return nil
}

func getNodeType(ctx context.Context, client *mongo.Client) (string, error) {
	md := proto.MasterDoc{}
	if err := client.Database("admin").RunCommand(ctx, primitive.M{"isMaster": 1}).Decode(&md); err != nil {
		return "", err
	}

	if md.SetName != nil || md.Hosts != nil {
		return TypeShardServer, nil
	} else if md.Msg == TypeIsDBGrid {
		// isdbgrid is always the msg value when calling isMaster on a mongos
		// see http://docs.mongodb.org/manual/core/sharded-cluster-query-router/
		return TypeMongos, nil
	}

	return TypeMongod, nil
}

func countMongodProcesses() (int, error) {
	pids, err := process.Pids()
	if err != nil {
		return 0, err
	}

	count := 0

	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		if name, _ := p.Name(); name == TypeMongod || name == TypeMongos {
			count++
		}
	}

	return count, nil
}
