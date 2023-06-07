package riverjson

// Various concrete types used to marshal River values.
type (
	// jsonStatement is a statement within a River body.
	jsonStatement interface{ isStatement() }

	// A jsonBody is a collection of statements.
	jsonBody = []jsonStatement

	// jsonBlock represents a River block as JSON. jsonBlock is a jsonStatement.
	jsonBlock struct {
		Name  string          `json:"name,omitempty"`
		Type  string          `json:"type,omitempty"` // Always "block"
		Label string          `json:"label,omitempty"`
		Body  []jsonStatement `json:"body,omitempty"`
	}

	// jsonAttr represents a River attribute as JSON. jsonAttr is a
	// jsonStatement.
	jsonAttr struct {
		Name  string    `json:"name,omitempty"`
		Type  string    `json:"type,omitempty"` // Always "attr"
		Value jsonValue `json:"value,omitempty"`
	}

	// jsonValue represents a single River value as JSON.
	jsonValue struct {
		Type  string      `json:"type,omitempty"`
		Value interface{} `json:"value,omitempty"`
	}

	// jsonObjectField represents a field within a River object.
	jsonObjectField struct {
		Key   string      `json:"key,omitempty"`
		Value interface{} `json:"value,omitempty"`
	}
)

func (jsonBlock) isStatement() {}
func (jsonAttr) isStatement()  {}
