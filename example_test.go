package schematics_test

import (
	"errors"
	"fmt"

	schematics "github.com/ashbeelghouri/json-schematics-v2"
)

// Example loads the bundled person schema and prints the validation errors for
// an invalid document. Errors are reported in schema field order.
func Example() {
	s := schematics.New()
	if err := s.LoadFile("examples/person.schema.json"); err != nil {
		panic(err)
	}

	data := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{"name": "a", "age": 200},
			"tags":    []any{"x", "x"},
		},
	}

	if err := s.Validate(data); err != nil {
		var ve *schematics.ValidationErrors
		if errors.As(err, &ve) {
			for _, msg := range ve.Strings("en", "%target: %message") {
				fmt.Println(msg)
			}
		}
	}
	// Output:
	// user.profile.name: name is too short
	// user.profile.age: age must be 0-120
	// user.tags: items must be unique (duplicate x)
}
