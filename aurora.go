package aurora

import (
	"bytes"
	"fmt"
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
	rgx := regexp.MustCompile(`(?is)(\{\{\s*for\.([a-z\-\_]*)\.in\.([a-z\-\_]*))(.*?)(endfor\s*\}\})`)
	var loopVars []interface{}
	for _, loop := range rgx.FindAllSubmatch(aurora.Templates[templateName].Data, -1) {
		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, loop[1], []byte(""), 1)
		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, loop[5], []byte(""), 1)

		re := regexp.MustCompile(`(?is)\{\{\s*` + string(loop[2]) + `\.{1}([a-z\_\-]*)\s*\}\}`)
		loopstr := ""

		arrData := datas[string(loop[3])].([]interface{})
		for _, item := range arrData {
			temp := loop[4]
			temp = re.ReplaceAllLiteral(temp, []byte("%v"))
			loopstr += string(temp)

			for _, d := range re.FindAllSubmatch(loop[4], -1) {
				if string(d[1]) == "" {
					loopVars = append(loopVars, item)
				} else {
					loopVars = append(loopVars, item.(map[string]interface{})[string(d[1])])
				}
			}
		}

		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, loop[4], []byte(loopstr), 1)
	}
	aurora.Templates[templateName].Data = []byte(fmt.Sprintf(string(aurora.Templates[templateName].Data), loopVars...))

	loopVars = nil
	rgx = regexp.MustCompile(`\{\{\s*([a-z]*)\s*\}\}`)
	for _, param := range rgx.FindAllSubmatch(aurora.Templates[templateName].Data, -1) {
		loopVars = append(loopVars, datas[string(param[1])])
		aurora.Templates[templateName].Data = bytes.Replace(aurora.Templates[templateName].Data, param[0], []byte("%v"), 1)
	}
	aurora.Templates[templateName].Data = []byte(fmt.Sprintf(string(aurora.Templates[templateName].Data), loopVars...))

	return aurora.Templates[templateName].Data
}
