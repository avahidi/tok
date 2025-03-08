package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSingle(t *testing.T) {
	for _, order := range []binary.ByteOrder{binary.BigEndian, binary.LittleEndian} {
		buf := new(bytes.Buffer)
		const INPUT1 = "this is a string"
		if err := WriteOne(buf, order, INPUT1); err != nil {
			t.Errorf("Failed to save string: %v", err)
		} else {
			var str string
			if err := ReadOne(buf, order, &str); err != nil {
				t.Errorf("Failed to load string: %v", err)
			} else if str != INPUT1 {
				t.Errorf("Incorrect string serialization: wanted '%s' got '%s'", INPUT1, str)
			}
		}
	}
}

func TestMultiple(t *testing.T) {
	for _, order := range []binary.ByteOrder{binary.BigEndian, binary.LittleEndian} {
		buf := new(bytes.Buffer)
		const INPUT1 = "this is a string"
		const INPUT2 uint32 = 111
		const INPUT3 = "this is another string"

		if err := WriteMultiple(buf, order, INPUT1, INPUT2, INPUT3); err != nil {
			t.Errorf("Failed to save multiple: %v", err)
		} else {
			var str1 string
			var int2 uint32
			var str3 string
			if err := ReadMultiple(buf, order, &str1, &int2, &str3); err != nil {
				t.Errorf("Failed to load multiple: %v", err)
			} else {
				if str1 != INPUT1 {
					t.Errorf("Incorrect multi-serialization (1): wanted '%s' got '%s'", INPUT1, str1)
				}
				if int2 != INPUT2 {
					t.Errorf("Incorrect multi-serialization (2): wanted '%d' got '%d'", INPUT2, int2)
				}
				if str3 != INPUT3 {
					t.Errorf("Incorrect multi-serialization (3): wanted '%s' got '%s'", INPUT3, str3)
				}
			}
		}
	}
}
