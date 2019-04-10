package jsonshape

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/steinarvk/orclib/lib/canonicalgojson"
)

func showPrimitive(shape Shape) ([]string, error) {
	short, ok, err := ShortSchemaDescription(shape)
	if err != nil {
		return nil, err
	}
	if ok {
		return []string{"<" + short + ">"}, nil
	}

	// fallback: encode the schema compactly
	data, err := canonicalgojson.MarshalCanonicalGoJSON(shape)
	if err != nil {
		return nil, err
	}

	return []string{string(data)}, nil
}

func Show(w io.Writer, shape Shape) error {
	lines, err := show(shape, "")
	if err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	return nil
}

func isObjectShape(shape Shape) bool {
	_, err := shapeFields(shape)
	return err == nil
}

func isArrayShape(shape Shape) bool {
	_, err := shapeItemShape(shape)
	return err == nil
}

func isAnyofShape(shape Shape) bool {
	_, err := anyofOptions(shape)
	return err == nil
}

func showLines(lines []string) string {
	if len(lines) == 1 {
		return " " + lines[0] + "\n"
	}
	rv := []string{""}
	indent := "  "
	for _, line := range lines {
		rv = append(rv, indent+line)
	}
	core := strings.TrimSpace(strings.Join(rv, "\n")) + "\n"
	if strings.HasPrefix(core, "#") {
		return "  " + core
	}
	return "\n" + indent + core
}

func show(shape Shape, fieldName string) ([]string, error) {
	w := &bytes.Buffer{}
	switch {
	case isObjectShape(shape):
		fields, err := shapeFields(shape)
		if err != nil {
			return nil, err
		}

		for _, field := range fields {
			if isArrayShape(field.Shape) {
				subshape, err := shapeItemShape(field.Shape)
				if err != nil {
					return nil, err
				}
				lines, err := show(subshape, field.Name)
				if err != nil {
					return nil, err
				}
				fmt.Fprintf(w, "%s[]:%s", field.Name, showLines(lines))
			} else {
				lines, err := show(field.Shape, field.Name)
				if err != nil {
					return nil, err
				}
				fmt.Fprintf(w, "%s:%s", field.Name, showLines(lines))
			}
		}

	case isArrayShape(shape):
		// array on the root, a bit obscure
		subshape, err := shapeItemShape(shape)
		if err != nil {
			return nil, err
		}
		lines, err := show(subshape, fieldName)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(w, "[]:%s", showLines(lines))

	case isAnyofShape(shape):
		desc, ok, err := ShortSchemaDescription(shape)
		if err != nil {
			return nil, err
		}
		if ok {
			return []string{desc}, nil
		}
		subshapes, err := anyofOptions(shape)
		if err != nil {
			return nil, err
		}
		comment := fmt.Sprintf("# %d options", len(subshapes))
		var lines []string
		for i, subshape := range subshapes {
			lineset, err := show(subshape, fieldName)
			if err != nil {
				return nil, err
			}
			if i > 0 {
				if fieldName != "" {
					lines = append(lines, fmt.Sprintf("--  # %s", fieldName))
				} else {
					lines = append(lines, "--")
				}
			}
			for _, line := range lineset {
				if line != "" {
					lines = append(lines, line)
				}
			}
		}
		return append([]string{comment}, lines...), nil

	default:
		return showPrimitive(shape)
	}

	return strings.Split(w.String(), "\n"), nil
}
