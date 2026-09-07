// Command genschema derives the saved-result JSON schema from the public types.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/stokaro/unswell"
)

type object = map[string]any

func main() {
	definitions := object{}
	root := schemaType(reflect.TypeFor[unswell.RunResult](), definitions)
	root["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	root["$defs"] = definitions
	root["title"] = "Unswell saved analysis " + unswell.SchemaVersion
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func schemaType(value reflect.Type, definitions object) object {
	switch value.Kind() {
	case reflect.Struct:
		name := value.String()
		if _, exists := definitions[name]; !exists {
			definitions[name] = schemaStruct(value, definitions)
		}
		return object{"$ref": "#/$defs/" + name}
	case reflect.Pointer:
		return object{"anyOf": []any{schemaType(value.Elem(), definitions), object{"type": "null"}}}
	case reflect.Slice:
		return object{"type": []string{"array", "null"}, "items": schemaType(value.Elem(), definitions)}
	case reflect.String:
		return object{"type": "string"}
	case reflect.Bool:
		return object{"type": "boolean"}
	case reflect.Int, reflect.Int64:
		return object{"type": "integer"}
	case reflect.Float64:
		return object{"type": "number"}
	case reflect.Invalid, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Complex64,
		reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.UnsafePointer:
		panic("unsupported contract type: " + value.String())
	}
	panic("unrecognized reflection kind")
}

func schemaStruct(value reflect.Type, definitions object) object {
	properties := object{}
	required := []string{}
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" || !field.IsExported() {
			continue
		}
		if name == "" {
			name = field.Name
		}
		property := schemaType(field.Type, definitions)
		if name == "schema_version" {
			property["const"] = unswell.SchemaVersion
		}
		properties[name] = property
		if options != "omitempty" {
			required = append(required, name)
		}
	}
	return object{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
