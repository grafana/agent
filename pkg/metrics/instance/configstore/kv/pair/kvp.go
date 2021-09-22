package pair

// KVP is a Key-Value Pair for a key in a kv.Client.
type KVP struct {
	Key string

	// Value should be deserialised through the Client's codec.
	// Nil indicates a deletion.
	Value interface{}
}
