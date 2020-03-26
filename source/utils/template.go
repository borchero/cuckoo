package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/markbates/pkger"
)

// PopulateTemplateWrite replaces values in a template file with Go's native template engine and
// writes the result to an output file.
func PopulateTemplateWrite(templateFile string, outputFile string, values interface{}) error {
	b, err := PopulateTemplate(templateFile, values)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outputFile, b, 0644)
}

// PopulateTemplate replaces values in a template file with Go's native template engine and
// returns the contents.
func PopulateTemplate(templateFile string, values interface{}) ([]byte, error) {
	file, err := pkger.Open("/utils/templates/" + templateFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s", err)
	}
	defer file.Close()

	source, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %s", err)
	}

	tmpl, err := template.New(templateFile).Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("Error compiling template: %s", err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, values)
	if err != nil {
		return nil, fmt.Errorf("Error population template: %s", err)
	}

	return buf.Bytes(), nil
}
