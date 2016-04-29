package template

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"

	"github.com/QubitProducts/bamboo/services/service"
)

func hasWeight(data map[string]int, key string) bool {
	_, exists := data[key]
	return exists
}

func getService(data map[string]service.Service, appId string) service.Service {
	serviceModel, _ := data[appId]
	return serviceModel
}

/*
	Returns string content of a rendered template
*/
func RenderTemplate(templateName string, templateContent string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"hasWeight":  hasWeight,
		"getService": getService,
		"Split":      strings.Split,
		"Contains":   strings.Contains,
		"Join":       strings.Join,
		"Replace":    strings.Replace,
		"ToUpper":    strings.ToUpper,
		"ToLower":    strings.ToLower,
		"Atoi":       strconv.Atoi,
		"Itoa":       strconv.Itoa,
	}

	tpl := template.Must(template.New(templateName).Funcs(funcMap).Parse(templateContent))

	strBuffer := new(bytes.Buffer)

	err := tpl.Execute(strBuffer, data)
	if err != nil {
		return "", err
	}

	return strBuffer.String(), nil
}
