package main

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"
)

// chatgptで生成

// OrderedMapSS is a custom type to handle map[string]string with ordered keys.
type OrderedMapSS struct {
	*orderedmap.OrderedMap
}

// NewOrderedMapSS creates a new instance of OrderedMapSS.
func NewOrderedMapSS() *OrderedMapSS {
	return &OrderedMapSS{OrderedMap: orderedmap.New()}
}

// UnmarshalJSON implements the json.Unmarshaler interface for OrderedMapSS.
func (o *OrderedMapSS) UnmarshalJSON(data []byte) error {
	temp := make(map[string]string)
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	o.OrderedMap = orderedmap.New()
	for key, value := range temp {
		o.Set(key, value)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface for OrderedMapSS.
func (o *OrderedMapSS) MarshalJSON() ([]byte, error) {
	mapData := make(map[string]string)
	for _, key := range o.Keys() {
		value, exists := o.Get(key)
		if exists {
			mapData[key] = value.(string)
		}
	}
	return json.Marshal(mapData)
}
