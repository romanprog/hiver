package hiver

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/op/go-logging"
	"github.com/romanprog/hiver/internal/config"
	"github.com/romanprog/hiver/internal/registry"
	"github.com/romanprog/hiver/pkg/template"
	"gopkg.in/yaml.v2"
)

var log = logging.MustGetLogger("hiver")

type hiverSpec struct {
	StackName  string                            `yaml:"stack,omitempty"`
	Registries []registry.Spec                   `yaml:"registries,omitempty"`
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

// Start cmd.
func Start() {

	readCommons(&mainHiverConfig)
	prepareHiverManifest(&mainHiverConfig)

	// Check all required data in hiver configuration.
	err := mainHiverConfig.Check()

	checkErr(err)
	registry.AuthList(mainHiverConfig.Registries)
	processSwarmPackages(&mainHiverConfig)

}

func processSwarmPackages(hConfig *hiverSpec) {
	var pkgList []*SwarmPackage
	for name := range hConfig.Packages {
		if config.Global.Packages.NeedServe(name) {
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
	log.Debugf("Read hiver manifest from file: %s", config.Global.MainConfig)

	// Apply commons options.
	log.Debug("Applying commons templating to hiver manifest.")
	commonTmp := make(map[string]interface{})
	commonTmp["commons"] = hConfig.Commons
	// Templated manifest data
	var parsedFile bytes.Buffer
	log.Debugf("Manifest: %s", config.Global.MainConfig)
	err := template.ExecTemplate(config.Global.MainConfig, &parsedFile, &commonTmp)
	checkErr(err)

	log.Debug("Parse hiver file.")
	err = yaml.UnmarshalStrict(parsedFile.Bytes(), &hConfig)
	checkErr(err)
	log.Debugf("Commons applied: \n%s", parsedFile.Bytes())

}

// Read files with common manifests. Parse yaml to main spec config.
func readCommons(hConfig *hiverSpec) {
	var commonsData [][]byte

	for _, FileName := range config.Global.CommonsConfigs {
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
