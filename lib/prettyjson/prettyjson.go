package prettyjson

import (
	"bytes"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func Format(data interface{}) string {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		logrus.Errorf("Error formatting JSON: %v", err)
	}
	return buf.String()
}
