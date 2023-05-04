package proto

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// docsExamined is renamed from nscannedObjects in 3.2.0
// https://docs.mongodb.com/manual/reference/database-profiler/#system.profile.docsExamined
type SystemProfile struct {
	AllUsers        []interface{} `bson:"allUsers"`
	Client          string        `bson:"client"`
	CursorExhausted bool          `bson:"cursorExhausted"`
	DocsExamined    int           `bson:"docsExamined"`
	NscannedObjects int           `bson:"nscannedObjects"`
	ExecStats       struct {
		Advanced                    int `bson:"advanced"`
		ExecutionTimeMillisEstimate int `bson:"executionTimeMillisEstimate"`
		InputStage                  struct {
			Advanced                    int    `bson:"advanced"`
			Direction                   string `bson:"direction"`
			DocsExamined                int    `bson:"docsExamined"`
			ExecutionTimeMillisEstimate int    `bson:"executionTimeMillisEstimate"`
			Filter                      struct {
				Date struct {
					Eq string `bson:"$eq"`
				} `bson:"date"`
			} `bson:"filter"`
			Invalidates  int    `bson:"invalidates"`
			IsEOF        int    `bson:"isEOF"`
			NReturned    int    `bson:"nReturned"`
			NeedTime     int    `bson:"needTime"`
			NeedYield    int    `bson:"needYield"`
			RestoreState int    `bson:"restoreState"`
			SaveState    int    `bson:"saveState"`
			Stage        string `bson:"stage"`
			Works        int    `bson:"works"`
		} `bson:"inputStage"`
		Invalidates  int    `bson:"invalidates"`
		IsEOF        int    `bson:"isEOF"`
		LimitAmount  int    `bson:"limitAmount"`
		NReturned    int    `bson:"nReturned"`
		NeedTime     int    `bson:"needTime"`
		NeedYield    int    `bson:"needYield"`
		RestoreState int    `bson:"restoreState"`
		SaveState    int    `bson:"saveState"`
		Stage        string `bson:"stage"`
		Works        int    `bson:"works"`
	} `bson:"execStats"`
	KeyUpdates   int `bson:"keyUpdates"`
	KeysExamined int `bson:"keysExamined"`
	Locks        struct {
		Collection struct {
			AcquireCount struct {
				R int `bson:"R"`
			} `bson:"acquireCount"`
		} `bson:"Collection"`
		Database struct {
			AcquireCount struct {
				R int `bson:"r"`
			} `bson:"acquireCount"`
		} `bson:"Database"`
		Global struct {
			AcquireCount struct {
				R int `bson:"r"`
			} `bson:"acquireCount"`
		} `bson:"Global"`
		MMAPV1Journal struct {
			AcquireCount struct {
				R int `bson:"r"`
			} `bson:"acquireCount"`
		} `bson:"MMAPV1Journal"`
	} `bson:"locks"`
	Millis             int       `bson:"millis"`
	Nreturned          int       `bson:"nreturned"`
	Ns                 string    `bson:"ns"`
	NumYield           int       `bson:"numYield"`
	Op                 string    `bson:"op"`
	Protocol           string    `bson:"protocol"`
	Query              bson.D    `bson:"query"`
	UpdateObj          bson.D    `bson:"updateobj"`
	Command            bson.D    `bson:"command"`
	OriginatingCommand bson.D    `bson:"originatingCommand"`
	ResponseLength     int       `bson:"responseLength"`
	Ts                 time.Time `bson:"ts"`
	User               string    `bson:"user"`
	WriteConflicts     int       `bson:"writeConflicts"`
}

func NewExampleQuery(doc SystemProfile) ExampleQuery {
	return ExampleQuery{
		Ns:                 doc.Ns,
		Op:                 doc.Op,
		Query:              doc.Query,
		Command:            doc.Command,
		OriginatingCommand: doc.OriginatingCommand,
		UpdateObj:          doc.UpdateObj,
	}
}

// ExampleQuery is a subset of SystemProfile
type ExampleQuery struct {
	Ns                 string `bson:"ns" json:"ns"`
	Op                 string `bson:"op" json:"op"`
	Query              bson.D `bson:"query,omitempty" json:"query,omitempty"`
	Command            bson.D `bson:"command,omitempty" json:"command,omitempty"`
	OriginatingCommand bson.D `bson:"originatingCommand,omitempty" json:"originatingCommand,omitempty"`
	UpdateObj          bson.D `bson:"updateobj,omitempty" json:"updateobj,omitempty"`
}

func (self ExampleQuery) Db() string {
	ns := strings.SplitN(self.Ns, ".", 2)
	if len(ns) > 0 {
		return ns[0]
	}
	return ""
}

// ExplainCmd returns bson.D ready to use in https://godoc.org/labix.org/v2/mgo#Database.Run
func (self ExampleQuery) ExplainCmd() bson.D {
	cmd := self.Command

	switch self.Op {
	case "query":
		if len(cmd) == 0 {
			cmd = self.Query
		}

		// MongoDB 2.6:
		//
		// "query" : {
		//   "query" : {
		//
		//   },
		//	 "$explain" : true
		// },
		if _, ok := cmd.Map()["$explain"]; ok {
			cmd = bson.D{
				{"explain", ""},
			}
			break
		}

		if len(cmd) == 0 || cmd[0].Key != "find" {
			var filter interface{}
			if len(cmd) > 0 && cmd[0].Key == "query" {
				filter = cmd[0].Value
			} else {
				filter = cmd
			}

			coll := ""
			s := strings.SplitN(self.Ns, ".", 2)
			if len(s) == 2 {
				coll = s[1]
			}

			cmd = bson.D{
				{"find", coll},
				{"filter", filter},
			}
		} else {
			for i := range cmd {
				switch cmd[i].Key {
				// PMM-1905: Drop "ntoreturn" if it's negative.
				case "ntoreturn":
					// If it's non-negative, then we are fine, continue to next param.
					if cmd[i].Value.(int64) >= 0 {
						continue
					}
					fallthrough
				// Drop $db as it is not supported in MongoDB 3.0.
				case "$db":
					if len(cmd)-1 == i {
						cmd = cmd[:i]
					} else {
						cmd = append(cmd[:i], cmd[i+1:]...)
					}
				}
			}
		}
	case "update":
		s := strings.SplitN(self.Ns, ".", 2)
		coll := ""
		if len(s) == 2 {
			coll = s[1]
		}
		if len(cmd) == 0 {
			cmd = bson.D{
				{Key: "q", Value: self.Query},
				{Key: "u", Value: self.UpdateObj},
			}
		}
		cmd = bson.D{
			{Key: "update", Value: coll},
			{Key: "updates", Value: []interface{}{cmd}},
		}
	case "remove":
		s := strings.SplitN(self.Ns, ".", 2)
		coll := ""
		if len(s) == 2 {
			coll = s[1]
		}
		if len(cmd) == 0 {
			cmd = bson.D{
				{Key: "q", Value: self.Query},
				// we can't determine if limit was 1 or 0 so we assume 0
				{Key: "limit", Value: 0},
			}
		}
		cmd = bson.D{
			{Key: "delete", Value: coll},
			{Key: "deletes", Value: []interface{}{cmd}},
		}
	case "insert":
		if len(cmd) == 0 {
			cmd = self.Query
		}
		if len(cmd) == 0 || cmd[0].Key != "insert" {
			coll := ""
			s := strings.SplitN(self.Ns, ".", 2)
			if len(s) == 2 {
				coll = s[1]
			}

			cmd = bson.D{
				{"insert", coll},
			}
		}
	case "getmore":
		if len(self.OriginatingCommand) > 0 {
			cmd = self.OriginatingCommand
			for i := range cmd {
				// drop $db param as it is not supported in MongoDB 3.0
				if cmd[i].Key == "$db" {
					if len(cmd)-1 == i {
						cmd = cmd[:i]
					} else {
						cmd = append(cmd[:i], cmd[i+1:]...)
					}
					break
				}
			}
		} else {
			cmd = bson.D{
				{Key: "getmore", Value: ""},
			}
		}
	case "command":
		cmd = sanitizeCommand(cmd)

		if len(cmd) == 0 || cmd[0].Key != "group" {
			break
		}

		if group, ok := cmd[0].Value.(bson.D); ok {
			for i := range group {
				// for MongoDB <= 3.2
				// "$reduce" : function () {}
				// It is then Unmarshaled as empty value, so in essence not working
				//
				// for MongoDB >= 3.4
				// "$reduce" : {
				//    "code" : "function () {}"
				// }
				// It is then properly Unmarshaled but then explain fails with "not code"
				//
				// The $reduce function shouldn't affect explain execution plan (e.g. what indexes are picked)
				// so we ignore it for now until we find better way to handle this issue
				if group[i].Key == "$reduce" {
					group[i].Value = "{}"
					cmd[0].Value = group
					break
				}
			}
		}
	}

	return bson.D{
		{
			Key:   "explain",
			Value: cmd,
		},
	}
}

func sanitizeCommand(cmd bson.D) bson.D {
	if len(cmd) < 1 {
		return cmd
	}

	key := cmd[0].Key
	if key != "count" && key != "distinct" {
		return cmd
	}

	for i := range cmd {
		// drop $db param as it is not supported in MongoDB 3.0
		if cmd[i].Key == "$db" {
			if len(cmd)-1 == i {
				cmd = cmd[:i]
			} else {
				cmd = append(cmd[:i], cmd[i+1:]...)
			}
			break
		}
	}

	return cmd
}
