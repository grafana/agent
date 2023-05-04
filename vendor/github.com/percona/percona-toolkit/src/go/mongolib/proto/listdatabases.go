package proto

import "go.mongodb.org/mongo-driver/bson/primitive"

// CollectionEntry represents an entry for ListCollections
type CollectionEntry struct {
	Name    string `bson:"name"`
	Type    string `bson:"type"`
	Options struct {
		Capped      bool  `bson:"capped"`
		Size        int64 `bson:"size"`
		AutoIndexID bool  `bson:"autoIndexId"`
	} `bson:"options"`
	Info struct {
		ReadOnly bool             `bson:"readOnly"`
		UUID     primitive.Binary `bson:"uuid"`
	} `bson:"info"`
}
