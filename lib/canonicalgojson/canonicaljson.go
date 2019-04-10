package canonicalgojson

import (
	"bytes"
	"encoding/json"
	"fmt"

	canonicaljson "github.com/docker/go/canonical/json"
	"github.com/steinarvk/orclib/lib/prettyjson"
)

func marshaller(data interface{}) ([]byte, error) {
	return canonicaljson.MarshalCanonical(data)
}

func MarshalCanonicalGoJSON(data interface{}) ([]byte, error) {
	canonicalized, err := marshaller(data)
	if err != nil {
		return nil, err
	}

	var sanityCheck interface{}
	err = json.Unmarshal(canonicalized, &sanityCheck)
	if err != nil {
		return nil, err
	}
	shouldBeEqual, err := marshaller(sanityCheck)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(canonicalized, shouldBeEqual) {
		return nil, fmt.Errorf("canonicalgojson failed sanity check on value:", prettyjson.Format(data))
	}

	return canonicalized, nil
}
