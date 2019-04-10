package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/steinarvk/orclib/lib/jsonshape"
	"github.com/steinarvk/orclib/lib/prettyjson"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		logrus.Fatal(err)
	}

	var structure interface{}
	if err := json.Unmarshal(data, &structure); err != nil {
		logrus.Fatal(err)
	}

	shape, err := jsonshape.ShapeOf(structure)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Println(prettyjson.Format(shape))

	if err := jsonshape.Show(os.Stdout, shape); err != nil {
		logrus.Fatal(err)
	}
}
