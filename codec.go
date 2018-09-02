package wsnet

import "encoding/json"

type (
	// Marshaler represents a marshal interface
	Marshaler interface {
		Marshal(interface{}) ([]byte, error)
	}

	// Unmarshaler represents a Unmarshal interface
	Unmarshaler interface {
		Unmarshal([]byte, interface{}) error
	}

	// Serializer is the interface that groups the basic Marshal and Unmarshal methods.
	Serializer interface {
		Marshaler
		Unmarshaler
	}
)

type jsonSerializer struct{}

// NewJSONSerializer returns a jsonSerializer
func NewJSONSerializer() Serializer {
	return &jsonSerializer{}
}

func (s *jsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
func (s *jsonSerializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
