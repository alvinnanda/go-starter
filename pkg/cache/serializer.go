package cache

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

// MarshalJSON serializes an object to JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON deserializes JSON into an object
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// MarshalGob serializes an object using gob encoding (more efficient than JSON for Go types)
func MarshalGob(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalGob deserializes gob-encoded data into an object
func UnmarshalGob(data []byte, v interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(v)
}
