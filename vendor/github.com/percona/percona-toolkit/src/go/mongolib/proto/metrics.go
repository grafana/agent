package proto

type Metrics struct {
	Commands      map[string]CommandStats `bson:"commands"`
	Cursor        *Cursor                 `bson:"cursor"`
	Document      *Document               `bson:"document"`
	GetLastError  *GetLastError           `bson:"getLastError"`
	Moves         float64                 `bson:"record.moves"`
	Operation     *Operation              `bson:"operation"`
	QueryExecutor *QueryExecutor          `bson:"queryExecutor"`
	Repl          *ReplMetrics            `bson:"repl"`
	Storage       *Storage                `bson:"storage"`
	Ttl           *Ttl                    `bson:"ttl"`
}

type CommandStats struct {
	Failed float64 `bson:"failed"`
	Total  float64 `bson:"total"`
}

type Cursor struct {
	NoTimeout float64 `bson:"open.noTimeout"`
	Pinned    float64 `bson:"open.pinned"`
	TimedOut  float64 `bson:"timedOut"`
	Total     float64 `bson:"open.total"`
}

type Document struct {
	Deleted  float64 `bson:"deleted"`
	Inserted float64 `bson:"inserted"`
	Returned float64 `bson:"returned"`
	Updated  float64 `bson:"updated"`
}

type GetLastError struct {
	Wtimeouts   float64 `bson:"wtimeouts"`
	Num         float64 `bson:"wtime.num"`
	TotalMillis float64 `bson:"wtime.totalMillis"`
}

type ReplMetrics struct {
	Batches            *MetricStats `bson:"apply.batches"`
	BufferSizeBytes    float64      `bson:"buffer.sizeBytes"`
	BufferCount        float64      `bson:"buffer.count"`
	BufferMaxSizeBytes float64      `bson:"buffer.maxSizeBytes"`
	Network            *ReplNetwork `bson:"network"`
	Ops                float64      `bson:"apply.ops"`
	PreloadDocs        *MetricStats `bson:"preload.docs"`
	PreloadIndexes     *MetricStats `bson:"preload.indexes"`
}

type Storage struct {
	BucketExhausted float64 `bson:"freelist.search.bucketExhausted"`
	Requests        float64 `bson:"freelist.search.requests"`
	Scanned         float64 `bson:"freelist.search.scanned"`
}

type MetricStats struct {
	Num         float64 `bson:"num"`
	TotalMillis float64 `bson:"totalMillis"`
}

type ReplNetwork struct {
	Getmores       *MetricStats `bson:"getmores"`
	Ops            float64      `bson:"ops"`
	ReadersCreated float64      `bson:"readersCreated"`
	Bytes          float64      `bson:"bytes"`
}

type Operation struct {
	Fastmod        float64 `bson:"fastmod"`
	Idhack         float64 `bson:"idhack"`
	ScanAndOrder   float64 `bson:"scanAndOrder"`
	WriteConflicts float64 `bson:"writeConflicts"`
}

type QueryExecutor struct {
	Scanned        float64 `bson:"scanned"`
	ScannedObjects float64 `bson:"scannedObjects"`
}

type Ttl struct {
	DeletedDocuments float64 `bson:"deletedDocuments"`
	Passes           float64 `bson:"passes"`
}
