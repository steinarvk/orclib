package jsonshape

import "testing"

func TestBasicSchemaDesc(t *testing.T) {
	testcases := struct {
		schema interface{}
		want   string
	}{
		{
			stringSchema,
			"string",
		},
		{
			numberSchema,
			"number",
		},
		{
			boolSchema,
			"bool",
		},
		{
			nullSchema,
			"null",
		},
		{
			arrayOfType(boolSchema),
			"[bool]",
		},
		{
			anyOf(numberSchema, stringSchema),
			"bool|string",
		},
		{
			anyOf(numberSchema, stringSchema, boolSchema),
			"bool|string?",
		},
		{
			objectWithProps(map[string]interface{}{
				"key": stringSchema,
			}),
			`{"key":string}`,
		},
	}

	for i, testcase := range testcases {
		got, ok, err := ShortSchemaDescription(testcase.schema)
		if !ok || err != nil {
			t.Errorf("#%d: ShortSchemaDescription(%v) = ok:%v, err:%v want %q", i, testcase.schema, ok, err, testcase.want)
		} else if got != testcase.want {
			t.Errorf("#%d: ShortSchemaDescription(%v) = %q want %q", i, testcase.schema, got, testcase.want)
		}
	}
}
