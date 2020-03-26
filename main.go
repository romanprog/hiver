package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// type nodeLabelsSpec struct {
// 	Name   string            `yaml:"name,omitempty"`
// 	Lables map[string]string `yaml:"labels,omitempty"`
// }

type hiverSpec struct {
	StackName  string                            `yaml:"stack,omitempty"`
	Registries []RegisrySpec                     `yaml:"registries,omitempty"`
	Packages   map[string]map[string]interface{} `yaml:"packages,omitempty"`

	//	Nodes      []nodeLabelsSpec                  `yaml:"nodes,omitempty"`
	Commons map[string]interface{}
}

// Check config.
// TODO: full check.
func (c *hiverSpec) Check() error {
	//log.Debugf("Stack name check: '%s'", c.StackName)
	if c.StackName == "" {
		return fmt.Errorf("stack name is empty. 'stack: stackname' field is required")
	}
	if len(c.Packages) == 0 {
		return errors.New("services count is 0. At least one is required")
	}
	return nil
}

// Parsed hiver manifest.
var mainHstackConfig hiverSpec

func main() {
	// Package main hiver.
	globalConfigInit()
	// Init logs module.
	loggingInit()
	// Init state (dotdir)
	// globalStateInit()
	//	colors()
	//return
	readCommons(&mainHstackConfig)
	readAndTmplManifest(&mainHstackConfig)

	// Check all required data in hiver configuration.
	err := mainHstackConfig.Check()

	checkErr(err)
	authRegistries(&mainHstackConfig)
	processSwarmPackages(&mainHstackConfig)
	//	tmplService("app1", &mainHstackConfig)
	//parseYamlManifest()

}

func processSwarmPackages(hconf *hiverSpec) {
	var pkgList []*SwarmPackage
	for name := range hconf.Packages {
		if globalConfig.Packages.NeedServe(name) {
			pkgList = append(pkgList, NewSwarmPackage(hconf, name))
		}
	}
	for _, pkg := range pkgList {
		pkg.ExecuteTemplate()
	}
	for _, pkg := range pkgList {
		pkg.Build()
	}
	for _, pkg := range pkgList {
		if pkg.IsInstalled() {
			pkg.DeploySwarm()
		} else {
			pkg.Delete()
		}
	}

}

func readAndTmplManifest(hconf *hiverSpec) {

	log.Info("Reading and parsing hiver manifest.")
	log.Debugf("Read hiver manifest from file: %s", globalConfig.MainConfig)

	// Apply commons options.
	log.Debug("Applying commons templating to hiver manifest.")
	commonTmp := make(map[string]interface{})
	commonTmp["commons"] = hconf.Commons
	// Templated manifest data
	var parsedFile bytes.Buffer
	err := ExecTemplate(globalConfig.MainConfig, &parsedFile, &commonTmp)
	checkErr(err)

	log.Debug("Parse hiver file.")
	err = yaml.UnmarshalStrict(parsedFile.Bytes(), &hconf)
	checkErr(err)
	log.Debugf("Commons applyed: \n%s", parsedFile.Bytes())

}

// Read files with common manifests. Parse yaml to main spec config.
func readCommons(hconf *hiverSpec) {
	var commonsData [][]byte

	for _, fname := range globalConfig.CommonsConfigs {
		log.Debugf("Read common yaml: %s", fname)
		tmpStr, err := ioutil.ReadFile(fname)
		checkErr(err)
		commonsData = append(commonsData, tmpStr)
	}
	for _, commonStr := range commonsData {
		err := yaml.Unmarshal(commonStr, &hconf.Commons)
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
