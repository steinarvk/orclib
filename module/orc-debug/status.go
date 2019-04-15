package orcdebug

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	canonicalhost "github.com/steinarvk/orclib/module/orc-canonicalhost"
	httprouter "github.com/steinarvk/orclib/module/orc-httprouter"
)

var (
	tableTemplate = template.Must(template.New("page").Parse(`
{{ if .TableName }}
	<h1>{{ .TableName }}</h1>
{{ end }}
<table>
	{{ range .Rows }}
		<tr>
			<td>{{ .Key }}
			<td>{{ .Value }}
	{{ end }}
</table>
`))
)

type Row struct {
	Key   string
	Value string
}

type Table struct {
	TableName string
	Rows      []Row
}

var startupTime time.Time

func init() {
	startupTime = time.Now()
}

type Status struct {
	tables []func() Table
}

func NewStatus() *Status {
	return &Status{
		tables: []func() Table{
			func() Table {
				return Table{
					TableName: "Overview",
					Rows: []Row{
						{"Startup time", fmt.Sprintf("%v", startupTime)},
						{"Uptime", fmt.Sprintf("%v", time.Since(startupTime))},
						{"Canonical host", canonicalhost.CanonicalHost},
					},
				}
			},
		},
	}
}

func (s *Status) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	w.Write([]byte("<html>\n"))

	for _, tableFunc := range s.tables {
		tableTemplate.Execute(w, tableFunc())
	}

	w.Write([]byte("<h1>/debug index</h1>"))
	trustedLinkListingTemplate.Execute(w, httprouter.M.ListDebugHandlers())

	w.Write([]byte("</html>\n"))
}

func (s *Status) AddTable(table func() Table) {
	s.tables = append(s.tables, table)
}
