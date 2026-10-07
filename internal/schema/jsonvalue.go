package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// orderedObject is a JSON object decoded while preserving the original key
// order, since encoding/json's map[string]interface{} does not.
type orderedObject struct {
	keys   []string
	values map[string]interface{}
}

func newOrderedObject() *orderedObject {
	return &orderedObject{values: make(map[string]interface{})}
}

func (o *orderedObject) set(key string, value interface{}) {
	if _, exists := o.values[key]; !exists {
		o.keys = append(o.keys, key)
	}
	o.values[key] = value
}

// ParseError describes a JSON parsing failure with a human-friendly
// location, so the UI can point the user at the approximate offending
// position.
type ParseError struct {
	Message string
	Line    int
	Column  int
	Offset  int64
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return e.Message
}

// decodeJSON parses raw JSON text into nested orderedObject/[]interface{}/
// scalar values, preserving object key order and distinguishing integers
// from floating point numbers.
func decodeJSON(raw []byte) (interface{}, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	value, err := decodeValue(dec)
	if err != nil {
		return nil, wrapDecodeError(err, raw, dec.InputOffset())
	}

	// Reject trailing garbage after the first value (e.g. "1 2" or two
	// concatenated objects), which json.Decoder otherwise ignores.
	if _, err := dec.Token(); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("unexpected content after the JSON value")
		}
		return nil, wrapDecodeError(err, raw, dec.InputOffset())
	}

	return value, nil
}

func decodeValue(dec *json.Decoder) (interface{}, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	return decodeTokenValue(dec, tok)
}

func decodeTokenValue(dec *json.Decoder, tok json.Token) (interface{}, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			obj := newOrderedObject()
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return nil, err
				}
				key, _ := keyTok.(string)
				if _, exists := obj.values[key]; exists {
					return nil, fmt.Errorf("duplicate JSON property %q", key)
				}

				val, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				obj.set(key, val)
			}
			if _, err := dec.Token(); err != nil { // consume closing '}'
				return nil, err
			}
			return obj, nil
		case '[':
			var arr []interface{}
			for dec.More() {
				val, err := decodeValue(dec)
				if err != nil {
					return nil, err
				}
				arr = append(arr, val)
			}
			if _, err := dec.Token(); err != nil { // consume closing ']'
				return nil, err
			}
			return arr, nil
		}
		return nil, fmt.Errorf("unexpected delimiter %q", t)
	case nil, bool, json.Number, string:
		return t, nil
	default:
		return nil, fmt.Errorf("unexpected token %v", t)
	}
}

// wrapDecodeError converts a json/encoding error into a ParseError with a
// best-effort line/column, falling back to the decoder's current read
// offset when the error itself does not carry one.
func wrapDecodeError(err error, raw []byte, fallbackOffset int64) error {
	offset := fallbackOffset
	message := err.Error()

	switch e := err.(type) {
	case *json.SyntaxError:
		offset = e.Offset
		message = e.Error()
	case *json.UnmarshalTypeError:
		offset = e.Offset
		message = e.Error()
	}

	if offset <= 0 {
		offset = fallbackOffset
	}

	line, col := lineAndColumn(raw, offset)
	return &ParseError{Message: message, Line: line, Column: col, Offset: offset}
}

// lineAndColumn converts a byte offset into a 1-based line/column pair.
func lineAndColumn(raw []byte, offset int64) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > int64(len(raw)) {
		offset = int64(len(raw))
	}

	line := 1
	col := 1
	for i := int64(0); i < offset; i++ {
		if raw[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
