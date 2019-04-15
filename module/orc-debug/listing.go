package orcdebug

import "html/template"

var (
	trustedLinkListingPageTemplate = template.Must(template.New("indexPage").Parse(`
<html>
	<body>
		{{ template "indexList" . }}
	</body>
</html>
`))

	trustedLinkListingTemplate = template.Must(template.New("indexList").Parse(`
		<ul>
			{{ range . }}
				<li> <a href="{{ . }}">{{ . }}</a>
			{{ end }}
		</ul>
`))
)
