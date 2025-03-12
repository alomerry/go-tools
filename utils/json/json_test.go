package json

import (
	"fmt"
	"testing"
)

func TestMarshal(t *testing.T) {
	json := Marshal(map[string]string{"a": "b"})
	fmt.Println(json)
}

func TestMarshalE(t *testing.T) {
	json, err := MarshalE(map[string]string{"a": "b"})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(json)
}
