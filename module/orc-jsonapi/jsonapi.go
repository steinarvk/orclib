package jsonapi

import (
	"fmt"

	"github.com/steinarvk/sectiontrace"
)

var APIPrefix = "/api/"

func (m *Module) Handle(endpoint string, methods Methods) {
	handler := wrapForStats(endpoint, methods.makeHandler())

	endpointSection := sectiontrace.New(fmt.Sprintf("JsonAPI(%s)", endpoint))
	handler = sectiontrace.WrapHandler(endpointSection, handler)

	m.apimux.Path(endpoint).Handler(handler)
}
