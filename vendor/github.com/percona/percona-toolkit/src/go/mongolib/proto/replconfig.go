package proto

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	ReplicaSetConfigTags map[string]string

	GetLastErrorModes map[string]*ReplicaSetConfigTags
)

// https://docs.mongodb.com/v3.2/reference/command/getLastError/#dbcmd.getLastError
type GetLastErrorDefaults struct {
	Journal      bool  `bson:"j,omitempty"`        // If true, wait for the next journal commit before returning, rather than waiting for a full disk flush.
	WriteConcern int64 `bson:"w,omitempty"`        // When running with replication, this is the number of servers to replicate to before returning.
	WTimeout     int64 `bson:"wtimeout,omitempty"` // Optional. Milliseconds. Specify a value in milliseconds to control how long to wait for write propagation to complete.
}

// https://docs.mongodb.com/v3.2/reference/replica-configuration/#rsconf.members
type ReplicaSetConfigMember struct {
	ID           int64                 `bson:"_id,omitempty"`          // An integer identifier of every member in the replica set.
	Host         string                `bson:"host,omitempty"`         // The hostname and, if specified, the port number, of the set member.
	ArbiterOnly  bool                  `bson:"arbiterOnly,omitempty"`  // A boolean that identifies an arbiter. A value of true indicates that the member is an arbiter.
	BuildIndexes bool                  `bson:"buildIndexes,omitempty"` // A boolean that indicates whether the mongod builds indexes on this member.
	Hidden       bool                  `bson:"hidden,omitempty"`       // When this value is true, the replica set hides this instance and does not include the member in the output of db.isMaster() or isMaster.
	Priority     int64                 `bson:"priority,omitempty"`     // A number that indicates the relative eligibility of a member to become a primary.
	Tags         *ReplicaSetConfigTags `bson:"tags,omitempty"`         // A tag set document containing mappings of arbitrary keys and values.
	SlaveDelay   int64                 `bson:"slaveDelay,omitempty"`   // The number of seconds “behind” the primary that this replica set member should “lag”.
	Votes        int64                 `bson:"votes,omitempty"`        // The number of votes a server will cast in a replica set election.
}

// https://docs.mongodb.com/v3.2/reference/replica-configuration/#rsconf.settings
type ReplicaSetConfigSettings struct {
	ChainingAllowed         bool                  `bson:"chainingAllowed,omitempty"`         // When chainingAllowed is true, the replica set allows secondary members to replicate from other secondary members.
	HeartbeatTimeoutSecs    int64                 `bson:"heartbeatTimeoutSecs,omitempty"`    // Number of seconds that the replica set members wait for a successful heartbeat from each other.
	HeartbeatIntervalMillis int64                 `bson:"heartbeatIntervalMillis,omitempty"` // The frequency in milliseconds of the heartbeats.
	ElectionTimeoutMillis   int64                 `bson:"electionTimeoutMillis,omitempty"`   // The time limit in milliseconds for detecting when a replica set’s primary is unreachable.
	GetLastErrorDefaults    *GetLastErrorDefaults `bson:"getLastErrorDefaults,omitempty"`    // A document that specifies the write concern for the replica set.
	GetLastErrorModes       *GetLastErrorModes    `bson:"getLastErrorModes,omitempty"`       // A document used to define an extended write concern through the use of members[n].tags.
	ReplicaSetId            *primitive.ObjectID   `bson:"replicaSetId,omitempty"`            // Replset Id (ObjectId)
}

type ReplicaSetConfig struct {
	Config struct {
		ID              string                    `bson:"_id,omitempty"`             // The name of the replica set. Once set, you cannot change the name of a replica set.
		ProtocolVersion int64                     `bson:"protocolVersion,omitempty"` // By default, new replica sets in MongoDB 3.2 use protocolVersion: 1. Previous versions of MongoDB use version 0.
		Version         int64                     `bson:"version,omitempty"`         // An incrementing number used to distinguish revisions of the replica set configuration object from previous iterations.
		Members         []*ReplicaSetConfigMember `bson:"members,omitempty"`         // An array of member configuration documents, one for each member of the replica set.
		Settings        *ReplicaSetConfigSettings `bson:"settings,omitempty"`        // A document that contains configuration options that apply to the whole replica set.
	} `bson:"config,omitempty"` // https://docs.mongodb.com/v3.2/reference/replica-configuration/#replica-set-configuration-fields
	Ok int64 `bson:"ok,omitempty"`
}
