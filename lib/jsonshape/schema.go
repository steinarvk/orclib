package jsonshape

import (
	"fmt"
	"sort"
	"strings"

	"github.com/steinarvk/orclib/lib/jsonwalk"
	"github.com/steinarvk/orclib/lib/orchash"
)

var (
	stringSchema = map[string]interface{}{
		"type": "string",
	}

	numberSchema = map[string]interface{}{
		"type": "number",
	}

	boolSchema = map[string]interface{}{
		"type": "boolean",
	}

	nullSchema = map[string]interface{}{
		"type": "null",
	}

	emptyArraySchema = map[string]interface{}{
		"type":     "array",
		"maxItems": float64(0),
	}
)

func hashSchema(schema interface{}) (string, error) {
	h, err := orchash.ComputeJSONHash(schema)
	if err != nil {
		return "", fmt.Errorf("Unhashable structure passed as schema: %v", schema)
	}
	return h, nil
}

func anyOf(schemas ...interface{}) interface{} {
	return map[string]interface{}{
		"anyOf": schemas,
	}
}

func arrayOfType(schema interface{}) interface{} {
	return map[string]interface{}{
		"type":  "array",
		"items": schema,
		//map[string]interface{}{
		//			"type": schema,
		//		},
	}
}

func objectWithProps(propSchemas map[string]interface{}) interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": propSchemas,
	}
}

func ShortSchemaDescription(x interface{}) (string, bool, error) {
	var isAnyOf bool
	var anyOfChoices []interface{}
	var itemsSchema interface{}
	var properties map[string]interface{}

	var maxItems *float64
	var isSimpleType string
	var isComplexType interface{}

	m, ok := x.(map[string]interface{})
	if !ok {
		return "", false, fmt.Errorf("Schema must be object, was %v", x)
	}
	numFields := len(m)

	err := jsonwalk.Walk(x, func(path string, rawValue interface{}) (bool, error) {
		if path == "" {
			return true, nil
		}
		switch {
		case path == ".anyOf":
			isAnyOf = true
			values, ok := rawValue.([]interface{})
			if !ok {
				return false, fmt.Errorf("%q must be array, was %v", ".anyOf", rawValue)
			}
			anyOfChoices = values
		case path == ".type":
			s, ok := rawValue.(string)
			if ok {
				isSimpleType = s
			} else {
				isComplexType = rawValue
			}
		case path == ".properties":
			values, ok := rawValue.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("%q must be array, was %v", ".properties", rawValue)
			}
			properties = values
		case path == ".items":
			itemsSchema = rawValue
		case path == ".maxItems":
			maxItemsValue, ok := rawValue.(float64)
			if !ok {
				return false, fmt.Errorf("%q must be number, was %v", ".properties", rawValue)
			}
			maxItems = &maxItemsValue
		default:
			return false, fmt.Errorf("Unknown field %q", path)
		}
		return false, nil
	})
	if err != nil {
		return "", false, err
	}

	switch {
	case isSimpleType == "object" && numFields == 2:
		if len(properties) == 0 {
			return "", false, fmt.Errorf("Expected properties for object")
		}
		if len(properties) > 1 {
			return "", false, nil
		}
		var key, valuetypedesc string
		for k, v := range properties {
			desc, ok, err := ShortSchemaDescription(v)
			if !ok || err != nil {
				return "", false, err
			}
			key = k
			valuetypedesc = desc
		}
		return fmt.Sprintf("{%q:%s}", key, valuetypedesc), true, nil

	case isSimpleType == "array" && numFields == 2 && maxItems != nil:
		if *maxItems != 0 {
			return "", false, nil
		}

		// empty array type
		return "[]", true, nil

	case isSimpleType == "array" && numFields == 2 && itemsSchema != nil:
		itemsDesc, ok, err := ShortSchemaDescription(itemsSchema)
		if !ok || err != nil {
			return "", false, err
		}

		return fmt.Sprintf("[%s]", itemsDesc), true, nil

	case isSimpleType != "" && numFields == 1:
		return isSimpleType, true, nil

	case isComplexType != nil && numFields == 1:
		return ShortSchemaDescription(isComplexType)

	case isAnyOf && numFields == 1 && len(anyOfChoices) > 0:
		var choicestrings []string
		seen := map[string]bool{}
		nullable := false
		for _, x := range anyOfChoices {
			choicestring, ok, err := ShortSchemaDescription(x)
			if !ok || err != nil {
				return "", false, err
			}
			if seen[choicestring] {
				continue
			}
			seen[choicestring] = true
			if choicestring == "null" {
				nullable = true
				continue
			}
			choicestrings = append(choicestrings, choicestring)
		}
		sort.Strings(choicestrings)
		rv := strings.Join(choicestrings, "|")
		suffix := ""
		if nullable {
			suffix = "?"
		}
		return fmt.Sprintf("%s%s", rv, suffix), true, nil

	default:
		return "", false, nil
	}
}
