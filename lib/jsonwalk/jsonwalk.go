package jsonwalk

import (
	"fmt"
	"reflect"
	"sort"
)

type Walker func(path string, value interface{}) (bool, error)

func Walk(structure interface{}, callback Walker, opts ...Option) error {
	optAccumulator := options{}
	for _, opt := range opts {
		if opt != nil {
			if err := opt(&optAccumulator); err != nil {
				return fmt.Errorf("Error setting option: %v", err)
			}
		}
	}

	return walk(optAccumulator, structure, callback)
}

func walk(opts options, structure interface{}, callback Walker) error {
	jsontype, err := TypeOf(structure)
	if err != nil {
		return fmt.Errorf("Error determining type of %v (Go type: %v): %v", structure, reflect.TypeOf(structure), err)
	}

	recurse, err := callback(opts.path, structure)
	if err != nil {
		return err
	}

	if !recurse {
		return nil
	}

	withPathSuffix := func(s string) options {
		rv := opts
		rv.path += s
		return rv
	}

	if jsontype == Object {
		m, err := objectAsMap(structure)
		if err != nil {
			return fmt.Errorf("Error processing %v (Go type: %v): unable to convert Object to map: %v", structure, reflect.TypeOf(structure), err)
		}

		ks := make([]string, len(m))
		var i int
		for k := range m {
			ks[i] = k
			i++
		}
		sort.Strings(ks)

		for _, k := range ks {
			val := m[k]
			suffix := "." + k
			if err := walk(withPathSuffix(suffix), val, callback); err != nil {
				return err
			}
		}
	} else if jsontype == Array {
		vals, err := arrayAsSlice(structure)
		if err != nil {
			return fmt.Errorf("Error processing %v (Go type: %v): unable to convert Array to slice: %v", structure, reflect.TypeOf(structure), err)
		}

		for i, val := range vals {
			var suffix string
			if opts.elideArrayIndices {
				suffix = "[]"
			} else {
				suffix = fmt.Sprintf("[%d]", i)
			}
			if err := walk(withPathSuffix(suffix), val, callback); err != nil {
				return err
			}
		}
	}

	return nil
}
