package proto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

const (
	REPLICA_SET_MEMBER_STARTUP = iota
	REPLICA_SET_MEMBER_PRIMARY
	REPLICA_SET_MEMBER_SECONDARY
	REPLICA_SET_MEMBER_RECOVERING
	REPLICA_SET_MEMBER_STARTUP2
	REPLICA_SET_MEMBER_UNKNOWN
	REPLICA_SET_MEMBER_ARBITER
	REPLICA_SET_MEMBER_DOWN
	REPLICA_SET_MEMBER_ROOLBACK
	REPLICA_SET_MEMBER_REMOVED
)

type Optime struct {
	Ts primitive.Timestamp `bson:"ts"` // The Timestamp of the last operation applied to this member of the replica set from the oplog.
	T  float64             `bson:"t"`  // The term in which the last applied operation was originally generated on the primary.
}

type Members struct {
	Optime               map[string]Optime   `bson:"optimes"`              // See Optime struct
	OptimeDate           primitive.DateTime  `bson:"optimeDate"`           // The last entry from the oplog that this member applied.
	InfoMessage          string              `bson:"infoMessage"`          // A message
	ID                   int64               `bson:"_id"`                  // Server ID
	Name                 string              `bson:"name"`                 // server name
	Health               float64             `bson:"health"`               // This field conveys if the member is up (i.e. 1) or down (i.e. 0).
	StateStr             string              `bson:"stateStr"`             // A string that describes state.
	Uptime               float64             `bson:"uptime"`               // number of seconds that this member has been online.
	ConfigVersion        float64             `bson:"configVersion"`        // revision # of the replica set configuration object from previous iterations of the configuration.
	Self                 bool                `bson:"self"`                 // true if this is the server we are currently connected
	State                float64             `bson:"state"`                // integer between 0 and 10 that represents the replica state of the member.
	ElectionTime         primitive.Timestamp `bson:"electionTime"`         // For the current primary, information regarding the election Timestamp from the operation log.
	ElectionDate         primitive.DateTime  `bson:"electionDate"`         // For the current primary, an ISODate formatted date string that reflects the election date.
	LastHeartbeat        primitive.DateTime  `bson:"lastHeartbeat"`        // Reflects the last time the server that processed the replSetGetStatus command received a response from a heartbeat that it sent to this member.
	LastHeartbeatRecv    primitive.DateTime  `bson:"lastHeartbeatRecv"`    // Reflects the last time the server that processed the replSetGetStatus command received a heartbeat request from this member.
	LastHeartbeatMessage string              `bson:"lastHeartbeatMessage"` // Contains a string representation of that message.
	PingMs               *float64            `bson:"pingMs,omitempty"`     // Represents the number of milliseconds (ms) that a round-trip packet takes to travel between the remote member and the local instance.
	Set                  string              `bson:"-"`
	StorageEngine        StorageEngine
}

// Struct for replSetGetStatus
type ReplicaSetStatus struct {
	Date                    primitive.DateTime `bson:"date"`                    // Current date
	MyState                 float64            `bson:"myState"`                 // Integer between 0 and 10 that represents the replica state of the current member
	Term                    float64            `bson:"term"`                    // The election count for the replica set, as known to this replica set member. Mongo 3.2+
	HeartbeatIntervalMillis float64            `bson:"heartbeatIntervalMillis"` // The frequency in milliseconds of the heartbeats. 3.2+
	Members                 []Members          `bson:"members"`                 //
	Ok                      float64            `bson:"ok"`                      //
	Set                     string             `bson:"set"`                     // Replica set name
}

type Member struct {
	Host         string  `bson:"host"`
	Votes        int32   `bson:"votes"`
	ID           int32   `bson:"_id"`
	SlaveDelay   int64   `bson:"slaveDelay"`
	Priority     float64 `bson:"priority"`
	BuildIndexes bool    `bson:"buildIndexes"`
	ArbiterOnly  bool    `bson:"arbiterOnly"`
	Hidden       bool    `bson:"hidden"`
	Tags         bson.M  `bson:"tags"`
}

type RSConfig struct {
	ID                                 string     `bson:"_id"`
	ConfigServer                       bool       `bson:"configsvr"`
	WriteConcernMajorityJournalDefault bool       `bson:"writeConcernMajorityJournalDefault"`
	Version                            int32      `bson:"version"`
	ProtocolVersion                    int64      `bson:"protocolVersion"`
	Settings                           RSSettings `bson:"settings"`
	Members                            []Member   `bson:"members"`
}

type LastErrorDefaults struct {
	W        interface{} `bson:"w"`
	WTimeout int32       `bson:"wtimeout"`
}

type RSSettings struct {
	HeartbeatTimeoutSecs       int32              `bson:"heartbeatTimeoutSecs"`
	ElectionTimeoutMillis      int32              `bson:"electionTimeoutMillis"`
	CatchUpTimeoutMillis       int32              `bson:"catchUpTimeoutMillis"`
	GetLastErrorModes          bson.M             `bson:"getLastErrorModes"`
	ChainingAllowed            bool               `bson:"chainingAllowed"`
	HeartbeatIntervalMillis    int32              `bson:"heartbeatIntervalMillis"`
	CatchUpTakeoverDelayMillis int32              `bson:"catchUpTakeoverDelayMillis"`
	GetLastErrorDefaults       LastErrorDefaults  `bson:"getLastErrorDefaults"`
	ReplicaSetID               primitive.ObjectID `bson:"replicaSetId"`
}

type Signature struct {
	Hash  primitive.Binary `bson:"hash"`
	KeyID int64            `bson:"keyId"`
}

type ClusterTime struct {
	ClusterTime primitive.Timestamp `bson:"clusterTime"`
	Signature   Signature           `bson:"signature"`
}

type ReplicasetConfig struct {
	Config              RSConfig            `bson:"config"`
	OK                  float64             `bson:"ok"`
	LastCommittedOpTime primitive.Timestamp `bson:"lastCommittedOpTime"`
	ClusterTime         ClusterTime         `bson:"$clusterTime"`
	OperationTime       primitive.Timestamp `bson:"operationTime"`
}

type ConfigVersion struct {
	ID                   int32              `bson:"_id"`
	MinCompatibleVersion int32              `bson:"minCompatibleVersion"`
	CurrentVersion       int32              `bson:"currentVersion"`
	ClusterID            primitive.ObjectID `bson:"clusterId"`
}

type ShardIdentity struct {
	ID                        string             `bson:"_id"`
	ShardName                 string             `bson:"shardName"`
	ClusterID                 primitive.ObjectID `bson:"clusterId"`
	ConfigsvrConnectionString string             `bson:"configsvrConnectionString"`
}

// MyState is a subset of getDiagnosticData result used to tag metrics in the MongoDB exporter
type MyState struct {
	Data struct {
		ReplicasetGetStatus struct {
			MyState int `bson:"myState"`
		} `bson:"replSetGetStatus"`
	} `bson:"data"`
}
