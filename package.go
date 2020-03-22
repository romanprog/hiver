package main

import (
	"bytes"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"gopkg.in/yaml.v2"
)

type SwarmPackageState struct {
	ManifestFile    string
	ManifestTmpFile string
}

type SwarmPackageBuildSpec struct {
	Enabled bool
	Dir     string `yaml:"dir,omitempty"`
	Type    string `yaml:"type,omitempty"`
	Script  string `yaml:"script,omitempty"`
	Args    string `yaml:"args,omitempty"`
	Check   bool   `yaml:check,omitempty`
	Image   string
	Tag     string
}

type SwarmPackage struct {
	configData     map[string]interface{}
	stack          string
	name           string
	dir            string
	readyForDeploy bool
	manifest       bytes.Buffer
	installed      bool
	build          SwarmPackageBuildSpec
	state          SwarmPackageState
}

func ParseBuild(b interface{}, spec *SwarmPackage) {

	if b == nil {
		log.Notice("Build spec is empty.")
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
			log.Fatalf("For build type 'dockerfile' field 'packages.%s.image' is requred.", spec.Name())
		}
		if spec.configData["tag"] == nil {
			log.Fatalf("For build type 'dockerfile' field 'packages.%s.tag' is requred.", spec.Name())
		}
		spec.build.Image = spec.configData["image"].(string)
		spec.build.Tag = spec.configData["tag"].(string)
	case "script":
		if spec.configData["script"] == nil {
			log.Fatalf("For build type 'dockerfile' field 'packages.%s.build.script' is requred.", spec.Name())
		}
	default:
		log.Panicf("Wrong build type: %s, suppurted only 'dockerfile|script' ")
	}
	if spec.configData["dir"] == nil {
		log.Fatalf("For build option, field 'packages.%s.build.dir' is requred.", spec.Name())
	}
	spec.build.Enabled = true
	return
}

func NewSwarmPackage(hconf *hiverSpec, name string) *SwarmPackage {
	pkg := new(SwarmPackage)
	log.Debugf("Initing swarm package '%s', check configuration.", name)
	if _, ok := hconf.Packages[name]; ok {
		pkg.configData = hconf.Packages[name]
		pkg.configData["commons"] = hconf.Commons
		pkg.name = name
		pkg.state.ManifestFile = filepath.Join(".", globalConfig.DotDir, "packages", name+".yaml")
		pkg.state.ManifestTmpFile = filepath.Join(".", globalConfig.DotDir, "packages", "_tmp_"+name+".yaml")
		err := os.MkdirAll(filepath.Dir(pkg.state.ManifestFile), os.ModePerm)
		checkErr(err)
		err = os.MkdirAll(filepath.Dir(pkg.state.ManifestTmpFile), os.ModePerm)
		checkErr(err)

		pkg.stack = hconf.StackName

		installed, err := strconv.ParseBool(pkg.configData["installed"].(string))
		if err != nil {
			log.Fatal("Field 'installed' is not set, but required.")
		}
		pkg.installed = installed
		pkgdir, ok := pkg.configData["dir"].(string)
		if !ok {
			log.Fatal("Field 'dir' is not set, but required.")
		}
		ParseBuild(pkg.configData["build"], pkg)
		pkg.dir = pkgdir
		log.Debugf("Package '%s' inited.", name)
		return pkg
	}
	log.Panicf("Can't create service unit. Service %s not found in hstack manifest.", name)
	return nil
}

func buildSwarmPackage(serviceName string, hstack *hiverSpec) {

	return
}

func (pkg *SwarmPackage) Tmpl() {

	if len(pkg.name) == 0 {
		log.Panicf("Service unit is not inited. Use NewService()")
	}

	pkgFname := filepath.Join(pkg.dir, globalConfig.SwarmPkgTmplFile)
	log.Debugf("Loading service tmplate: %s", pkgFname)

	tplFile, err := ioutil.ReadFile(pkgFname)
	checkErr(err)

	log.Infof("Templating service: %s", pkg.name)
	tmpl, err := template.New(pkg.name).Option("missingkey=error").Parse(string(tplFile))
	checkErr(err)
	err = tmpl.Execute(&pkg.manifest, pkg.configData)
	log.Debugf("Service %s manifest after templating: \n%s", pkg.name, pkg.manifest.String())
	checkErr(err)
	pkg.readyForDeploy = true

}

// Deploy package to swarm.
func (pkg *SwarmPackage) DeploySwarm() {
	pkg.SaveStateTmp()
	man, err := ioutil.ReadFile(pkg.state.ManifestFile)
	diff, err := diffYamls(man, pkg.ManifestBuff().Bytes(), true)
	checkErr(err)

	if globalConfig.DryRun {
		log.Noticef("Dry run: Package '%s', manifests diff: \n%s", pkg.name, diff)
		return
	}

	deployCommand := fmt.Sprintf("docker stack deploy -c %s --with-registry-auth %s", pkg.state.ManifestTmpFile, pkg.stack)
	log.Infof("Deploying manifest '%s'", pkg.Name())
	err = commandExec(deployCommand)

	checkErr(err)
	pkg.SaveState()
}

// Get list of docker services from pakage manifest.
func (pkg *SwarmPackage) Delete() {
	pkg.SaveStateTmp()
	list := getPkgServices(pkg.manifest.Bytes())
	for _, sname := range list {
		log.Infof("Deleting package '%s', service '%s'", pkg.Name(), sname)
		deleteCommand := fmt.Sprintf("docker service rm %s_%s", pkg.stack, sname)
		cmdres, cmderr, err := commandExecOutput(deleteCommand)
		if err != nil {
			log.Debugf("Deleting output: %s; Error (it's normal for this operation) %s", cmdres, cmderr)
			return
		}
		log.Debugf("Deleting output: %s", cmdres)
	}
}

func (pkg *SwarmPackage) Manifest() string {
	res := pkg.manifest.String()
	return res

}

func (pkg *SwarmPackage) BuildDockerfile() {

	if !pkg.build.Enabled {
		log.Debug("Skip build for package '%s'", pkg.name)
		return
	}
	// Check if image exists in registry. (To skip extra builds)
	if pkg.build.Check {
		checkCommand := fmt.Sprintf("docker pull %s:%s", pkg.build.Image, pkg.build.Tag)
		log.Infof("Pulling image '%s:%s'", pkg.build.Image, pkg.build.Tag)
		err := commandExec(checkCommand)
		if err == nil {
			if globalConfig.DryRun {
				log.Noticef("Dry run: image exists: '%s:%s', skiping build.", pkg.build.Image, pkg.build.Tag)
				return
			}
			log.Debugf("Image '%s:%s' exists. Skiping build.", pkg.build.Image, pkg.build.Tag)
			return
		}
	}
	if globalConfig.DryRun {
		log.Noticef("Dry run: build image: '%s:%s', build dir: '%s'", pkg.build.Image, pkg.build.Tag, pkg.build.Dir)
		return
	}
	// Build image.
	buildCommand := fmt.Sprintf("docker build -t %s:%s %s", pkg.build.Image, pkg.build.Tag, pkg.build.Dir)
	log.Infof("Building dockerfile '%s'", pkg.Name())
	err := commandExec(buildCommand)
	checkErr(err)

	// Push to registry.
	buildCommand = fmt.Sprintf("docker push %s:%s", pkg.build.Image, pkg.build.Tag)
	log.Infof("Pushing docker image '%s:%s'", pkg.build.Image, pkg.build.Tag)
	err = commandExec(buildCommand)
	checkErr(err)

}

func (pkg *SwarmPackage) Name() string {
	res := pkg.name
	return res

}

func (pkg *SwarmPackage) ManifestBuff() *bytes.Buffer {
	res := &pkg.manifest
	return res

}

func (pkg *SwarmPackage) SaveState() {
	savePackageSatate(pkg, false)
}

func (pkg *SwarmPackage) SaveStateTmp() {
	savePackageSatate(pkg, true)
}

func (pkg *SwarmPackage) IsInstalled() bool {
	return pkg.installed
}

func savePackageSatate(pkg *SwarmPackage, tmp bool) {
	if !pkg.readyForDeploy {
		log.Panic("Can's save manifest to file. Use init and template first.")
	}
	var fname string
	if tmp {
		fname = pkg.state.ManifestTmpFile
	} else {
		fname = pkg.state.ManifestFile
	}
	//	log.Debug("Creating state dir for packages. %s", filepath.Dir(pkg.SaveState()))
	log.Debugf("Saving manifest to file '%s'", fname)
	mf, err := os.Create(fname)
	checkErr(err)
	_, err = mf.Write(pkg.manifest.Bytes())
	checkErr(err)

}

func getPkgServices(manifest []byte) []string {
	var slist struct {
		Services map[string]interface{} `yaml:"services"`
	}
	err := yaml.Unmarshal(manifest, &slist)
	res := []string{}
	for sname, _ := range slist.Services {
		res = append(res, sname)
	}
	checkErr(err)
	return res
}
