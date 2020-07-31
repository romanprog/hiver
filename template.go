package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var funcMap template.FuncMap
var tmplTargetFileName string

// TmplFunctionsMap - return list of functions for adding to template.
func TmplFunctionsMap() template.FuncMap {
	funcMap := template.FuncMap{
		"envOrDef": envOrDefault,
		"env":      env,
		"fileMD5":  fileMD5,
		"Iterate":  Iterate,
	}
	return funcMap
}

// ExecTemplate apply template data 'data' to template file 'filename'.
// Write result to 'result'.
func ExecTemplate(filename string, result io.Writer, data interface{}) error {

	// Set curent target file name for function fileMD5
	tmplTargetFileName = filename
	// Read template from file.
	tmplData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	// Parse template.
	tmpl, err := template.New("main").Funcs(TmplFunctionsMap()).Option("missingkey=error").Parse(string(tmplData))
	if err != nil {
		return err
	}
	// Apply data to template.
	err = tmpl.Execute(result, &data)
	if err != nil {
		return err
	}
	return nil
}

// Helper for template.
// Get env or return deafult.
func envOrDefault(key string, defaultVal string) (string, error) {
	if envVal, ok := os.LookupEnv(key); ok {
		log.Debugf("ENV variable %s. Value %s", key, envVal)
		return envVal, nil
	}
	log.Debugf("ENV variable %s undefined, set default value: %s", key, defaultVal)
	return defaultVal, nil
}

// Function for template. Use: {{ env }}
// Get env, return error if env variable is not exist.
func env(key string) (string, error) {
	if envVal, ok := os.LookupEnv(key); ok {
		log.Debugf("ENV variable %s. Value %s", key, envVal)
		return envVal, nil
	}
	log.Debugf("Error, ENV variable %s undefined but needed", key)
	return string(""), fmt.Errorf("ENV variable %s undefined but needed", key)
}

// fileMD5 is function for template. Use: {{ fileMD5 "filename" size }}
// The path to the file should be relative to the template in which the function was called.
// Return file md5 checksum truncated to size bytes.
func fileMD5(filename string, sz int) (string, error) {

	fn := filepath.Join(filepath.Dir(tmplTargetFileName), filename)
	f, err := os.Open(fn)
	checkErr(err)
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)[0:sz]), nil
}

// stringMD5 is function for template. Use: {{ stringMD5 "string" size }}
// Return string md5 checksum truncated to size bytes.
func stringMD5(data string, sz int) (string, error) {

	h := md5.New()
	if _, err := io.WriteString(h, data); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)[0:sz]), nil
}

// Iterate - function for template.
// Implements a "like for loop" in templates.
// Use:
// {{- range $val := Iterate 5 }}
//   {{ $val }}
// {{- end }}
func Iterate(count uint) []uint {
	var i uint
	var items []uint
	for i = 0; i < count; i++ {
		items = append(items, i)
	}
	return items
}
