package hiver

import (
	"bytes"

	"github.com/romanprog/hiver/pkg/diff"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/romanprog/hiver/internal/runner"
	"github.com/romanprog/hiver/pkg/template"

	"github.com/romanprog/hiver/internal/config"
	"gopkg.in/yaml.v2"
)

// SwarmPackageState is type for state files names.
type SwarmPackageState struct {
	ManifestFile    string
	ManifestTmpFile string
}

// SwarmPackageBuildSpec is type for package build spec.
type SwarmPackageBuildSpec struct {
	Enabled   bool
	Dir       string `yaml:"dir,omitempty"`
	Type      string `yaml:"type,omitempty"`
	Script    string `yaml:"script,omitempty"`
	Args      string `yaml:"args,omitempty"`
	Check     bool   `yaml:"check,omitempty"`
	PushImage bool   `yaml:"push,omitempty"`
	Image     string
	Tag       string
}

// SwarmPackage is main package spec.
type SwarmPackage struct {
	configData     map[string]interface{} // Reference to parsed yaml struct.
	stack          string                 // Stack name.
	name           string                 // Package name.
	dir            string                 // Path to package.
	readyForDeploy bool                   // Package is parsed and inited.
	manifest       bytes.Buffer           // Manifest data (after templating).
	installed      bool                   // Option 'installed'.
	build          SwarmPackageBuildSpec  // Build spec.
	state          SwarmPackageState      // State files.
}

// ParseBuild -
// parse and check build data, create build spec.
func ParseBuild(b interface{}, spec *SwarmPackage) {

	if b == nil {
		log.Debugf("Build spec is empty for package '%s'.", spec.name)
		return
	}
	yamlData, err := yaml.Marshal(b)
	checkErr(err)
	err = yaml.UnmarshalStrict(yamlData, &spec.build)
	checkErr(err)

	// Check build spec.
	switch spec.build.Type {
	case "dockerfile":
		if spec.configData["image"] == nil {
			log.Fatalf("For build type 'dockerfile' field 'packages.%s.image' is required.", spec.Name())
		}
		if spec.configData["tag"] == nil {
			log.Fatalf("For build type 'dockerfile' field 'packages.%s.tag' is required.", spec.Name())
		}
	case "script":
		if spec.build.Script == "" {
			log.Fatalf("For build type 'script' field 'packages.%s.build.script' is required.", spec.Name())
		}
	default:
		log.Fatalf("Wrong build type '%s', supported only 'dockerfile|script'", spec.build.Type)
	}
	spec.build.Image, _ = spec.configData["image"].(string)
	spec.build.Tag, _ = spec.configData["tag"].(string)
	if spec.configData["dir"] == nil {
		log.Fatalf("For build option, field 'packages.%s.build.dir' is required.", spec.Name())
	}
	spec.build.Enabled = true
	return
}

// NewSwarmPackage create, check and init package speck.
func NewSwarmPackage(hSpec *hiverSpec, name string) *SwarmPackage {
	pkg := new(SwarmPackage)
	log.Debugf("Initing swarm package '%s', check configuration.", name)
	// Check if data for this name exists in main hiver manifest.
	if _, ok := hSpec.Packages[name]; !ok {
		log.Fatalf("Can't create service unit. Service %s not found in hiver manifest.", name)
	}

	pkg.configData = hSpec.Packages[name]
	pkg.configData["commons"] = hSpec.Commons
	pkg.name = name
	pkg.state.ManifestFile = filepath.Join(".", config.Global.DotDir, "packages", name+".yaml")
	err := os.MkdirAll(filepath.Dir(pkg.state.ManifestFile), os.ModePerm)
	checkErr(err)

	pkg.stack = hSpec.StackName

	ok := false
	pkg.installed, ok = pkg.configData["installed"].(bool)
	if !ok {
		log.Fatal("Field 'installed' is not set, but required.")
	}

	pkgdir, ok := pkg.configData["dir"].(string)
	if !ok {
		log.Fatal("Field 'dir' is not set, but required.")
	}

	// Parse build data.
	ParseBuild(pkg.configData["build"], pkg)

	pkg.dir = pkgdir
	pkg.state.ManifestTmpFile = filepath.Join(".", pkg.dir, "_tmp_"+name+".yaml")
	log.Debugf("Package '%s' inited.", name)
	return pkg

}

func buildSwarmPackage(serviceName string, hStack *hiverSpec) {

	return
}

// ExecuteTemplate - execute app templates for package.
// Save result to 'manifest' variable.
func (pkg *SwarmPackage) ExecuteTemplate() {

	if len(pkg.name) == 0 {
		log.Panicf("Service unit is not inited. Use NewService()")
	}

	pkgFilename := filepath.Join(pkg.dir, config.Global.SwarmPkgTmplFile)

	log.Debugf("Templating service: %s", pkg.name)
	err := template.ExecTemplate(pkgFilename, &pkg.manifest, pkg.configData)
	checkErr(err)
	log.Debugf("Service %s manifest after templating: \n%s", pkg.name, pkg.manifest.String())
	checkErr(err)
	pkg.readyForDeploy = true

}

// DeploySwarm - deploy package to swarm.
func (pkg *SwarmPackage) DeploySwarm() {
	pkg.SaveStateTmp()
	defer pkg.DeleteStateTmpl()
	man, err := ioutil.ReadFile(pkg.state.ManifestFile)
	diff, err := diff.Yamls(man, pkg.ManifestBuff().Bytes(), true)
	checkErr(err)

	if config.Global.DryRun {
		log.Noticef("Dry run: Package '%s', manifests diff: \n%s", pkg.name, diff)
		return
	}

	deployCommand := fmt.Sprintf("docker stack deploy -c %s --with-registry-auth %s", pkg.state.ManifestTmpFile, pkg.stack)
	log.Infof("Deploying manifest '%s'", pkg.Name())
	err = runner.BashExec(deployCommand)

	checkErr(err)
	pkg.SaveState()
}

// Delete - get list of docker services from package manifest
// and delete them.
func (pkg *SwarmPackage) Delete() {
	pkg.SaveStateTmp()
	list := getPkgServices(pkg.manifest.Bytes())
	for _, packageName := range list {
		if config.Global.DryRun {
			log.Noticef("Dry run: Package '%s', service '%s' will be deleted", pkg.Name(), packageName)
			continue
		}
		log.Infof("Deleting package '%s', service '%s'", pkg.Name(), packageName)
		deleteCommand := fmt.Sprintf("docker service rm %s_%s", pkg.stack, packageName)
		commandStdout, commandStderr, err := runner.BashExecOutput(deleteCommand)
		if err != nil {
			log.Debugf("Deleting output: %s; Error (it's normal for this operation) %s", commandStdout, commandStderr)
			return
		}
		log.Debugf("Deleting output: %s", commandStdout)

	}
}

// Manifest return pkg manifest file (after applying templates) as string.
func (pkg *SwarmPackage) Manifest() string {
	res := pkg.manifest.String()
	return res

}

// Build - check package, and run build of some "type".
func (pkg *SwarmPackage) Build() {
	if !pkg.installed {
		log.Debugf("Skip build for '%s', installed: false", pkg.name)
		return
	}
	if !pkg.build.Enabled {
		log.Debugf("Skip build for package '%s'", pkg.name)
		return
	}
	switch pkg.build.Type {
	case "dockerfile":
		pkgBuildDockerfile(pkg)
	case "script":
		pkgBuildScript(pkg)
	default:
		return
	}
}

// pkgBuildDockerfile - build package using dockerfile.
// Push image to repo (uses image and tag options)
func pkgBuildDockerfile(pkg *SwarmPackage) {

	// Check if image exists in registry. (To skip extra builds)
	if pkg.build.Check {
		checkCommand := fmt.Sprintf("docker pull %s:%s", pkg.build.Image, pkg.build.Tag)
		log.Infof("Pulling image '%s:%s'", pkg.build.Image, pkg.build.Tag)
		err := runner.BashExec(checkCommand)
		if err == nil {
			if config.Global.DryRun {
				log.Noticef("Dry run: image exists: '%s:%s', skiping build.", pkg.build.Image, pkg.build.Tag)
				return
			}
			log.Debugf("Image '%s:%s' exists. Skiping build.", pkg.build.Image, pkg.build.Tag)
			return
		}
	}
	// Check dry-run option.
	if config.Global.DryRun {
		log.Noticef("Dry run: build image: '%s:%s', build dir: '%s'", pkg.build.Image, pkg.build.Tag, pkg.build.Dir)
	} else {
		// Build image.
		buildCommand := fmt.Sprintf("docker build -t %s:%s %s", pkg.build.Image, pkg.build.Tag, pkg.build.Dir)
		log.Infof("Building package %s (dockerfile)", pkg.Name())
		err := runner.BashExec(buildCommand)
		checkErr(err)
	}
	if pkg.build.PushImage {
		return
	}
	if config.Global.DryRun {
		log.Noticef("Dry run: push image: '%s:%s'", pkg.build.Image, pkg.build.Tag)
		return
	}
	// Push to registry.
	buildCommand := fmt.Sprintf("docker push %s:%s", pkg.build.Image, pkg.build.Tag)
	log.Infof("Pushing docker image '%s:%s'", pkg.build.Image, pkg.build.Tag)
	err := runner.BashExec(buildCommand)
	checkErr(err)
}

// isFile check is 'filename' is really file.
func isFile(filename string) error {
	info, err := os.Stat(filename)

	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("'%s' is a directory", filename)
	}

	return nil
}

// pkgBuildDockerfile - build package using script.
// Runs script with image and tag as first and second arguments.
func pkgBuildScript(pkg *SwarmPackage) {

	scriptFilename := filepath.Join(pkg.build.Dir, pkg.build.Script)
	err := isFile(scriptFilename)
	checkErr(err)

	// Build image.
	buildCommand := fmt.Sprintf("cd %s && ./%s %s %s %s", pkg.build.Dir, pkg.build.Script, pkg.build.Image, pkg.build.Tag, pkg.build.Args)
	if config.Global.DryRun {
		log.Noticef("Dry run: build script: '%s', build dir: '%s'", pkg.build.Script, pkg.build.Dir)
		log.Debugf("Build command: \n%s", buildCommand)
		return
	}
	log.Infof("Building package '%s' (script run)", pkg.Name())
	err = runner.BashExec(buildCommand)
	checkErr(err)

}

// Name - return package name.
func (pkg *SwarmPackage) Name() string {
	res := pkg.name
	return res

}

// ManifestBuff - return package manifest as bytes buffer.
func (pkg *SwarmPackage) ManifestBuff() *bytes.Buffer {
	res := &pkg.manifest
	return res

}

// SaveState - save package manifest (with all applied templates) to file as state.
// Uses after successful deploy.
func (pkg *SwarmPackage) SaveState() {
	savePackageSatate(pkg, false)
}

// SaveStateTmp - save package manifest (with all applied templates) to tmp.
// Uses for deploy ().
func (pkg *SwarmPackage) SaveStateTmp() {
	savePackageSatate(pkg, true)
}

// DeleteStateTmpl - remove tmp manifest.
// Uses for deploy ().
func (pkg *SwarmPackage) DeleteStateTmpl() error {
	err := os.Remove(pkg.state.ManifestTmpFile)
	return err
}

// IsInstalled return option 'installed'.
func (pkg *SwarmPackage) IsInstalled() bool {
	return pkg.installed
}

func savePackageSatate(pkg *SwarmPackage, tmp bool) {
	if !pkg.readyForDeploy {
		log.Panic("Can's save manifest to file. Use init and template first.")
	}
	var fileName string
	if tmp {
		fileName = pkg.state.ManifestTmpFile
	} else {
		fileName = pkg.state.ManifestFile
	}
	//	log.Debug("Creating state dir for packages. %s", filepath.Dir(pkg.SaveState()))
	log.Debugf("Saving manifest to file '%s'", fileName)
	mf, err := os.Create(fileName)
	checkErr(err)
	_, err = mf.Write(pkg.manifest.Bytes())
	checkErr(err)

}

func getPkgServices(manifest []byte) []string {
	var servicesList struct {
		Services map[string]interface{} `yaml:"services"`
	}
	err := yaml.Unmarshal(manifest, &servicesList)
	res := []string{}
	for serviceName := range servicesList.Services {
		res = append(res, serviceName)
	}
	checkErr(err)
	return res
}
