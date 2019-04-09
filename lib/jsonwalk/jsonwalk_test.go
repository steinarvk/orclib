package jsonwalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

const (
	testJSONData = `
{
  "traceEvents": [
    {"cat": "Section", "scope": "testscope", "id": 1,
     "pid": 123,
     "name": "section1", "ph": "b",
     "ts": 1230000000
    },
    {"cat": "Section", "scope": "testscope", "id": 2,
     "pid": 123,
     "name": "section2", "ph": "b",
     "ts": 1231000000,
     "args": {"p": 1, "a": 1}
    },
    {"cat": "Section", "scope": "testscope", "id": 3,
     "pid": 123,
     "name": "section3", "ph": "b",
     "ts": 1234000000,
     "args": {"p": 2, "a": 1}
    },
    {"cat": "Section", "scope": "testscope", "id": 3,
     "pid": 123,
     "name": "section3", "ph": "e",
     "ts": 1244000000,
     "args": {"p": 2, "a": 1, "ok": true}
    },
    {"cat": "Section", "scope": "testscope", "id": 2,
     "pid": 123,
     "name": "section2", "ph": "e",
     "ts": 1248000000,
     "args": {"p": 1, "a": 1, "ok": false}
    },
    {"cat": "Section", "scope": "testscope", "id": 4,
     "pid": 123,
     "name": "section2", "ph": "b",
     "ts": 1250000000,
     "args": {"p": 1, "a": 1}
    },
    {"cat": "Section", "scope": "testscope", "id": 4,
     "pid": 123,
     "name": "section2", "ph": "e",
     "ts": 1255000000,
     "args": {"p": 1, "a": 1, "ok": true}
    },
    {"cat": "Section", "scope": "testscope", "id": 1,
     "pid": 123,
     "name": "section1", "ph": "e",
     "ts": 1259000000,
     "args": {"ok": true}
    }
  ],
  "displayTimeUnit": "ms",
  "otherData": {
    "test": "yes"
  }
}
`
)

func TestBasicWalking(t *testing.T) {
	var structure interface{}
	if err := json.Unmarshal([]byte(testJSONData), &structure); err != nil {
		logrus.Fatal(err)
	}

	buf := &bytes.Buffer{}

	if err := Walk(structure, func(path string, value interface{}) (bool, error) {
		switch castVal := value.(type) {
		case nil:
			fmt.Fprintf(buf, "%q: null\n", path)
		case string:
			fmt.Fprintf(buf, "%q: %s (%q)\n", path, "string", castVal)
		case float64:
			fmt.Fprintf(buf, "%q: %s (%s)\n", path, "number", strconv.FormatFloat(castVal, 'f', -1, 64))
		case bool:
			fmt.Fprintf(buf, "%q: %s (%v)\n", path, "bool", castVal)
		case map[string]interface{}:
			fmt.Fprintf(buf, "%q: object\n", path)
		case []interface{}:
			fmt.Fprintf(buf, "%q: array\n", path)
		default:
			return false, fmt.Errorf("Unknown kind of value: %v", value)
		}
		return true, nil
	}); err != nil {
		t.Fatal(err)
	}

	want := strings.TrimSpace(`
"": object
".displayTimeUnit": string ("ms")
".otherData": object
".otherData.test": string ("yes")
".traceEvents": array
".traceEvents[0]": object
".traceEvents[0].cat": string ("Section")
".traceEvents[0].id": number (1)
".traceEvents[0].name": string ("section1")
".traceEvents[0].ph": string ("b")
".traceEvents[0].pid": number (123)
".traceEvents[0].scope": string ("testscope")
".traceEvents[0].ts": number (1230000000)
".traceEvents[1]": object
".traceEvents[1].args": object
".traceEvents[1].args.a": number (1)
".traceEvents[1].args.p": number (1)
".traceEvents[1].cat": string ("Section")
".traceEvents[1].id": number (2)
".traceEvents[1].name": string ("section2")
".traceEvents[1].ph": string ("b")
".traceEvents[1].pid": number (123)
".traceEvents[1].scope": string ("testscope")
".traceEvents[1].ts": number (1231000000)
".traceEvents[2]": object
".traceEvents[2].args": object
".traceEvents[2].args.a": number (1)
".traceEvents[2].args.p": number (2)
".traceEvents[2].cat": string ("Section")
".traceEvents[2].id": number (3)
".traceEvents[2].name": string ("section3")
".traceEvents[2].ph": string ("b")
".traceEvents[2].pid": number (123)
".traceEvents[2].scope": string ("testscope")
".traceEvents[2].ts": number (1234000000)
".traceEvents[3]": object
".traceEvents[3].args": object
".traceEvents[3].args.a": number (1)
".traceEvents[3].args.ok": bool (true)
".traceEvents[3].args.p": number (2)
".traceEvents[3].cat": string ("Section")
".traceEvents[3].id": number (3)
".traceEvents[3].name": string ("section3")
".traceEvents[3].ph": string ("e")
".traceEvents[3].pid": number (123)
".traceEvents[3].scope": string ("testscope")
".traceEvents[3].ts": number (1244000000)
".traceEvents[4]": object
".traceEvents[4].args": object
".traceEvents[4].args.a": number (1)
".traceEvents[4].args.ok": bool (false)
".traceEvents[4].args.p": number (1)
".traceEvents[4].cat": string ("Section")
".traceEvents[4].id": number (2)
".traceEvents[4].name": string ("section2")
".traceEvents[4].ph": string ("e")
".traceEvents[4].pid": number (123)
".traceEvents[4].scope": string ("testscope")
".traceEvents[4].ts": number (1248000000)
".traceEvents[5]": object
".traceEvents[5].args": object
".traceEvents[5].args.a": number (1)
".traceEvents[5].args.p": number (1)
".traceEvents[5].cat": string ("Section")
".traceEvents[5].id": number (4)
".traceEvents[5].name": string ("section2")
".traceEvents[5].ph": string ("b")
".traceEvents[5].pid": number (123)
".traceEvents[5].scope": string ("testscope")
".traceEvents[5].ts": number (1250000000)
".traceEvents[6]": object
".traceEvents[6].args": object
".traceEvents[6].args.a": number (1)
".traceEvents[6].args.ok": bool (true)
".traceEvents[6].args.p": number (1)
".traceEvents[6].cat": string ("Section")
".traceEvents[6].id": number (4)
".traceEvents[6].name": string ("section2")
".traceEvents[6].ph": string ("e")
".traceEvents[6].pid": number (123)
".traceEvents[6].scope": string ("testscope")
".traceEvents[6].ts": number (1255000000)
".traceEvents[7]": object
".traceEvents[7].args": object
".traceEvents[7].args.ok": bool (true)
".traceEvents[7].cat": string ("Section")
".traceEvents[7].id": number (1)
".traceEvents[7].name": string ("section1")
".traceEvents[7].ph": string ("e")
".traceEvents[7].pid": number (123)
".traceEvents[7].scope": string ("testscope")
".traceEvents[7].ts": number (1259000000)
`)

	got := strings.TrimSpace(buf.String())

	if got != want {
		t.Errorf("Got flattened JSON:\n%s\nWanted flattened JSON:\n%s\n", got, want)
	}
}
