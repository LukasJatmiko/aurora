package aurora

import (
	"bytes"
	"io/ioutil"
	"log"
	"regexp"
)

//Aurora :
type Aurora struct {
	TemplatePath string
	RunMode      string
	Templates    map[string]*Template
}

//Template :
type Template struct {
	Name string
	Data []byte
}

//NewAurora :
func NewAurora(templatePath string, runMode string) *Aurora {
	rgx := regexp.MustCompile(`(\/)+$`)
	templatePath = rgx.ReplaceAllLiteralString(templatePath, "")
	return &Aurora{
		TemplatePath: templatePath,
		RunMode:      runMode,
		Templates:    make(map[string]*Template),
	}
}

//Init :
func (aurora *Aurora) Init() {
	if files, err := ioutil.ReadDir(aurora.TemplatePath); err != nil {
		log.Println(err)
	} else {
		isAuroraFile := regexp.MustCompile(`\.aurora$`)
		for _, file := range files {
			if isAuroraFile.MatchString(file.Name()) {
				templateName := isAuroraFile.ReplaceAllLiteralString(file.Name(), "")
				aurora.Templates[templateName] = &Template{Name: templateName}
				if aurora.Templates[templateName].Data, err = ioutil.ReadFile(aurora.TemplatePath + "/" + file.Name()); err != nil {
					log.Printf("Could not read template file %v (%v)\n", file.Name(), err)
				}
			}
		}
	}
}

//Render :
func (aurora *Aurora) Render(templateName string, datas map[string]interface{}) []byte {

	//render loops
	rgx := regexp.MustCompile(`(?is)(\{\{\s*for\.([a-z]*)\.in\.([a-z]*))(.*?)(endfor\s*\}\})`)
	for _, loop := range rgx.FindAllSubmatch(aurora.Templates[templateName].Data, -1) {
		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, loop[1], []byte(""), 1)
		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, loop[5], []byte(""), 1)

		re := regexp.MustCompile(`(?is)\{\{\s*` + string(loop[2]) + `\s*\}\}`)
		loopstr := ""
		for _, item := range datas[string(loop[3])].([]string) {
			temp := loop[4]
			temp = re.ReplaceAllLiteral([]byte(temp), []byte(item))
			loopstr += string(temp)
		}

		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, loop[4], []byte(loopstr), 1)
	}

	rgx = regexp.MustCompile(`\{\{\s*([a-z]*)\s*\}\}`)
	for _, param := range rgx.FindAllSubmatch(aurora.Templates[templateName].Data, -1) {
		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, param[0], []byte(datas[string(param[1])].(string)), 1)
	}

	return aurora.Templates[templateName].Data
}
