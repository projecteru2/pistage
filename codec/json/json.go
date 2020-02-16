package json

import "encoding/json"

// Decode .
func Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Encode .
func Encode(v interface{}, indents ...string) ([]byte, error) {
	var indent string
	if len(indents) > 0 {
		indent = indents[0]
	}
	return json.MarshalIndent(v, "", indent)
}
