package handle

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	loginTmpl   = template.Must(template.New("login").ParseFS(templateFS, "templates/login.html"))
	successTmpl = template.Must(template.New("success").ParseFS(templateFS, "templates/success.html"))
	errorTmpl   = template.Must(template.New("error").ParseFS(templateFS, "templates/error.html"))
)
