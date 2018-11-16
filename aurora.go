package aurora

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"sync"
)

var templatesync sync.RWMutex
var TemplateDir string
var Templates map[string][]byte

func Init() {
	Templates = make(map[string][]byte)
	TemplateDir = "./views"
}

//Render : Render template
func Render(templatefilepath string, renderobject map[string]interface{}) string {
	var err error
	rgx := regexp.MustCompile(`(?i)^[\/]*`)
	templatefilepath = rgx.ReplaceAllLiteralString(templatefilepath, "")
	templatesync.Lock()
	if len(Templates[templatefilepath]) < 1 { //parse template if template file isn't loaded
		Templates[templatefilepath], err = ioutil.ReadFile(TemplateDir + "/" + templatefilepath + ".aurora")
		templatesync.Unlock()
		if err != nil {
			fmt.Printf(`Error : %v`, err.Error())
			return "<!Doctype html><html><head><title>Error!</title></head><body><h1>Error while parsing template file.</h2></body></html>"
		}
	} else {
		templatesync.Unlock()
	}
	templatesync.Lock()
	html := Templates[templatefilepath]
	templatesync.Unlock()
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
