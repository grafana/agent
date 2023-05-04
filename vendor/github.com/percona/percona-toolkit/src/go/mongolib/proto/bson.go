package proto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BsonD bson.D

func (d *BsonD) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	t, err := dec.Token()
	if err != nil {
		return err
	}
	if t != json.Delim('{') {
		return fmt.Errorf("expected { but got %s", t)
	}
	for {
		t, err := dec.Token()
		if err != nil {
			return err
		}

		// Might be empty object
		if t == json.Delim('}') {
			return nil
		}

		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expected key to be a string but got %s", t)
		}

		de := primitive.E{}
		de.Key = key

		if !dec.More() {
			return fmt.Errorf("missing value for key %s", key)
		}

		var raw json.RawMessage
		err = dec.Decode(&raw)
		if err != nil {
			return err
		}

		var v BsonD
		err = bson.UnmarshalExtJSON(raw, true, &v)
		if err != nil {
			var v []BsonD
			err = bson.UnmarshalExtJSON(raw, true, &v)
			if err != nil {
				var v interface{}
				err = bson.UnmarshalExtJSON(raw, true, &v)
				if err != nil {
					return err
				} else {
					de.Value = v
				}
			} else {
				de.Value = v
			}
		} else {
			de.Value = v
		}

		*d = append(*d, de)
		if !dec.More() {
			break
		}
	}

	t, err = dec.Token()
	if err != nil {
		return err
	}
	if t != json.Delim('}') {
		return fmt.Errorf("expect delimeter %s but got %s", json.Delim('}'), t)
	}

	return nil
}

func (d BsonD) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer

	b.WriteByte('{')

	for i, v := range d {
		if i > 0 {
			b.WriteByte(',')
		}

		// marshal key
		key, err := bson.MarshalExtJSON(v.Key, false, true)
		if err != nil {
			return nil, err
		}
		b.Write(key)
		b.WriteByte(':')

		var val []byte
		if value, ok := v.Value.(float64); ok && math.IsInf(value, 0) {
			if math.IsInf(value, 1) {
				val = []byte("Infinity")
			} else {
				val = []byte("-Infinity")
			}

			// below is wrong, but I'm later unable to Unmarshal Infinity,
			// so we turn it into string for now
			val = append([]byte(`"`), val...)
			val = append(val, '"')
		} else {
			// marshal value
			val, err = bson.MarshalExtJSON(v.Value, false, true)
			if err != nil {
				return nil, err
			}
		}
		b.Write(val)
	}

	b.WriteByte('}')

	return b.Bytes(), nil
}

func (d BsonD) Len() int {
	return len(d)
}

// Map returns a map out of the ordered element name/value pairs in d.
func (d BsonD) Map() (m bson.M) {
	m = make(bson.M, len(d))
	for _, item := range d {
		switch v := item.Value.(type) {
		case BsonD:
			m[item.Key] = v.Map()
		case []BsonD:
			el := []bson.M{}
			for i := range v {
				el = append(el, v[i].Map())
			}
			m[item.Key] = el
		case []interface{}:
			// mgo/bson doesn't expose UnmarshalBSON interface
			// so we can't create custom bson.Unmarshal()
			el := []bson.M{}
			for i := range v {
				if b, ok := v[i].(BsonD); ok {
					el = append(el, b.Map())
				}
			}
			m[item.Key] = el
		default:
			m[item.Key] = item.Value
		}
	}
	return m
}
