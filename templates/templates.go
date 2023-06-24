package templates

type HtmlName struct {
	Home         string
	Dashboard    string
	Authenticate string
}

var TemplateEndpoints = "./templates"

var VIEWS = &HtmlName{
	Home:         TemplateEndpoints + "/index.html",
	Dashboard:    TemplateEndpoints + "/dashboard.html",
	Authenticate: TemplateEndpoints + "/authenticate.html",
}
