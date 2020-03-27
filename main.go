package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type hiverSpec struct {
	StackName  string                            `yaml:"stack,omitempty"`
	Registries []RegistrySpec                    `yaml:"registries,omitempty"`
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
		return errors.New("packages count is 0. At least one is required")
	}
	return nil
}

// Parsed hiver manifest.
var mainHiverConfig hiverSpec

func main() {
	// Package main hiver.
	loggingInit()
	testGit()

	return

	globalConfigInit()

	// Init logs module.
	loggingInit()
	readCommons(&mainHiverConfig)
	prepareHiverManifest(&mainHiverConfig)

	// Check all required data in hiver configuration.
	err := mainHiverConfig.Check()

	checkErr(err)
	authRegistries(&mainHiverConfig)
	processSwarmPackages(&mainHiverConfig)

}

func processSwarmPackages(hConfig *hiverSpec) {
	var pkgList []*SwarmPackage
	for name := range hConfig.Packages {
		if globalConfig.Packages.NeedServe(name) {
			pkgList = append(pkgList, NewSwarmPackage(hConfig, name))
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

func prepareHiverManifest(hConfig *hiverSpec) {

	log.Info("Reading and parsing hiver manifest.")
	log.Debugf("Read hiver manifest from file: %s", globalConfig.MainConfig)

	// Apply commons options.
	log.Debug("Applying commons templating to hiver manifest.")
	commonTmp := make(map[string]interface{})
	commonTmp["commons"] = hConfig.Commons
	// Templated manifest data
	var parsedFile bytes.Buffer
	err := ExecTemplate(globalConfig.MainConfig, &parsedFile, &commonTmp)
	checkErr(err)

	log.Debug("Parse hiver file.")
	err = yaml.UnmarshalStrict(parsedFile.Bytes(), &hConfig)
	checkErr(err)
	log.Debugf("Commons applied: \n%s", parsedFile.Bytes())

}

// Read files with common manifests. Parse yaml to main spec config.
func readCommons(hConfig *hiverSpec) {
	var commonsData [][]byte

	for _, FileName := range globalConfig.CommonsConfigs {
		log.Debugf("Read common yaml: %s", FileName)
		tmpStr, err := ioutil.ReadFile(FileName)
		checkErr(err)
		commonsData = append(commonsData, tmpStr)
	}
	for _, commonStr := range commonsData {
		err := yaml.Unmarshal(commonStr, &hConfig.Commons)
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
