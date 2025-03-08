package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

var (
	BYTE_ORDER = binary.BigEndian
)

// ReadExact is a helper function for reading an exact amount of bytes
func ReadExact(r io.Reader, size int) ([]byte, error) {
	buffer := make([]byte, size)
	n, err := r.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("Expected to read %d bytes, got %d", size, n)
	}
	return buffer, nil
}

// WriteExact is a helper function for writing an exact amount of bytes
func WriteExact(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("Expected to write %d bytes, got %d", len(data), n)
	}
	return nil
}

// ReadSized is a  helper function for reading data in form of <uint32 size><size bytes>
func ReadSized(r io.Reader, order binary.ByteOrder) ([]byte, error) {
	var size uint32
	if err := binary.Read(r, order, &size); err != nil {
		return nil, err
	}

	return ReadExact(r, int(size))
}

// WriteSizedis a helper function for writing data in form of <uint32 size><size bytes>
func WriteSized(w io.Writer, order binary.ByteOrder, data []byte) error {
	if err := binary.Write(w, order, uint32(len(data))); err != nil {
		return err
	}
	return WriteExact(w, data)
}

// ReadOne is a binary Read function that also supports strings
func ReadOne(r io.Reader, order binary.ByteOrder, obj any) error {
	switch x := obj.(type) {
	case *string:
		data, err := ReadSized(r, order)
		if err != nil {
			return err
		}
		*x = string(data)

		return nil
	default:
		return binary.Read(r, order, obj)
	}

}

// ReadMultiple is similar to ReadOne but allows multiple items to be read
func ReadMultiple(r io.Reader, order binary.ByteOrder, objs ...any) error {
	for _, obj := range objs {
		if err := ReadOne(r, order, obj); err != nil {
			return err
		}
	}
	return nil
}

// WriteOne is a binary Write function that allows certain other types
func WriteOne(w io.Writer, order binary.ByteOrder, obj any) error {
	switch x := obj.(type) {
	case string:
		asBytes := []byte(x)
		return WriteSized(w, order, asBytes)
	case []byte:
		if err := binary.Write(w, order, uint32(len(x))); err != nil {
			return err
		}
		_, err := w.Write(x)
		return err
	default:
		return binary.Write(w, order, obj)
	}
}

// WriteMultiple is similar to WriteOne but allows multiple items to be written
func WriteMultiple(w io.Writer, order binary.ByteOrder, objs ...any) error {
	for _, obj := range objs {
		if err := WriteOne(w, order, obj); err != nil {
			return err
		}
	}
	return nil
}
