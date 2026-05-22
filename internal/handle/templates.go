package handle

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	loginTmpl   = template.Must(template.ParseFS(templateFS, "templates/login.html"))
	successTmpl = template.Must(template.ParseFS(templateFS, "templates/success.html"))
	errorTmpl   = template.Must(template.ParseFS(templateFS, "templates/error.html"))
)
