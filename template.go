package main

import (
	//	"bytes"
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

func TmplFunctionsMap() template.FuncMap {
	funcMap := template.FuncMap{
		"envOrDef": envOrDefault,
		"env":      env,
		"fileMD5":  fileMD5,
		"testfunc": testFunc,
	}
	return funcMap
}

func TmplApply(filename string, templated io.Writer, commonTmpl interface{}) error {

	tmplTargetFileName = filename
	manifestData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	tmpl, err := template.New("main").Funcs(TmplFunctionsMap()).Option("missingkey=error").Parse(string(manifestData))
	if err != nil {
		return err
	}
	err = tmpl.Execute(templated, &commonTmpl)
	if err != nil {
		return err
	}
	//log.Error(templated.String())
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

// Get env, return error if env variable is not exist.
func env(key string) (string, error) {
	if envVal, ok := os.LookupEnv(key); ok {
		log.Debugf("ENV variable %s. Value %s", key, envVal)
		return envVal, nil
	}
	log.Debugf("Error. ENV variable %s undefined but needed.", key)
	return string(""), fmt.Errorf("Error. ENV variable %s undefined but needed.", key)
}

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

func stringMD5(data string, sz int) (string, error) {

	h := md5.New()
	if _, err := io.WriteString(h, data); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)[0:sz]), nil
}

func testFunc() (string, error) {
	return filepath.Dir("../../" + tmplTargetFileName), nil
}
