package jsonshape

type jsonSchema struct {
	Type       interface{} `json:"type,omitempty"`
	AnyOf      interface{} `json:"anyOf,omitempty"`
	Items      interface{} `json:"items,omitempty"`
	MaxItems   interface{} `json:"maxItems,omitempty"`
	Properties interface{} `json:"properties,omitempty"`
}
