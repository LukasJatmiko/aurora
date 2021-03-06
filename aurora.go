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
	composeView := aurora.Templates[templateName].Data
	rgx := regexp.MustCompile(`(?is)(\{\{\s*for\.([a-z\-\_]*)\.in\.([a-z\-\_]*))(.*?)(endfor\s*\}\})`)
	var loopVars []interface{}
	for _, loop := range rgx.FindAllSubmatch(composeView, -1) {
		composeView = bytes.Replace(composeView, loop[1], []byte(""), 1)
		composeView = bytes.Replace(composeView, loop[5], []byte(""), 1)

		re := regexp.MustCompile(`(?is)\{\{\s*` + string(loop[2]) + `\.{1}([a-z\_\-]*)\s*\}\}`)
		temp := loop[4]
		temp = re.ReplaceAllLiteral(temp, []byte("%v"))
		loopstr := ""

		switch datas[string(loop[3])].(type) {
		default:
			{
				for _, item := range datas[string(loop[3])].([]interface{}) {
					loopstr += string(temp)
					loopVars = append(loopVars, item)
				}
			}
		case []map[string]interface{}:
			{
				for _, item := range datas[string(loop[3])].([]map[string]interface{}) {
					loopstr += string(temp)
					for _, d := range re.FindAllSubmatch(loop[4], -1) {
						loopVars = append(loopVars, item[string(d[1])])
					}
				}
			}
		}

		composeView = bytes.Replace(composeView, loop[4], []byte(loopstr), 1)
	}
	composeView = []byte(fmt.Sprintf(string(composeView), loopVars...))

	loopVars = nil
	rgx = regexp.MustCompile(`(?is)\{\{\s*([a-z\-\_]*)\s*\}\}`)
	for _, param := range rgx.FindAllSubmatch(composeView, -1) {
		loopVars = append(loopVars, datas[string(param[1])])
		composeView = bytes.Replace(composeView, param[0], []byte("%v"), 1)
	}
	return []byte(fmt.Sprintf(string(composeView), loopVars...))
}
