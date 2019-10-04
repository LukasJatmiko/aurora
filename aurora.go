package aurora

import (
	"io/ioutil"
	"regexp"
	"sync"
)

//Aurora :
type Aurora struct {
	TemplateDir string
	Templates   map[string][]byte
	Sync        sync.RWMutex
}

//Options :
type Options struct {
	TemplateDir string
}

//Init :
func (a *Aurora) Init(opt *Options) {
	a.Templates = make(map[string][]byte)
	a.TemplateDir = opt.TemplateDir
}

//Render : Render template
func (a *Aurora) Render(templatefilepath string, renderobject map[string]interface{}) string {
	var err error
	rgx := regexp.MustCompile(`(?i)^[\/]*`)
	templatefilepath = rgx.ReplaceAllLiteralString(templatefilepath, "")
	a.Sync.Lock()
	if len(a.Templates[templatefilepath]) < 1 { //parse template if template file isn't loaded
		a.Templates[templatefilepath], err = ioutil.ReadFile(a.TemplateDir + "/" + templatefilepath + ".aurora")
		a.Sync.Unlock()
		if err != nil {
			return "<!Doctype html><html><head><title>Error!</title></head><body><h1>Error while parsing template file.</h2></body></html>"
		}
	} else {
		a.Sync.Unlock()
	}
	a.Sync.Lock()
	html := a.Templates[templatefilepath]
	a.Sync.Unlock()
	for key, value := range renderobject {
		rpattern := regexp.MustCompile(`(?i)\[\{\{` + key + `\}\}\]*`)
		if value != nil {
			html = rpattern.ReplaceAllLiteral(html, []byte(value.(string)))
		} else {
			html = rpattern.ReplaceAllLiteral(html, []byte(""))
		}
	}
	return string(html)
}
