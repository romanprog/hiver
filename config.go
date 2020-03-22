package main

import (
	"flag"
	"os"
	"path/filepath"
)

// Configuration.
type ConfSpec struct {
	Packages         packagesList
	Build            bool
	Debug            bool
	MainConfig       string
	DryRun           bool
	CommonsConfigs   argsArrayType
	DotDir           string
	SwarmPkgTmplFile string
	WorkDir          string
}

type argsArrayType []string
type packagesList struct {
	argsArrayType
}

// Method for argsArrayType.
func (i *argsArrayType) String() string {
	return "my string representation"
}

// Method for argsArrayType.
func (i *argsArrayType) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Search in args list.
func (i *argsArrayType) Find(value string) bool {
	if len(*i) == 0 {
		return false
	}
	for _, elem := range *i {
		if value == elem {
			return true
		}
	}
	return false
}

// Search service name, return true in case of empty list (all services options).
func (i *packagesList) NeedServe(value string) bool {
	if len(i.argsArrayType) == 0 {
		return true
	}
	return i.Find(value)
}

// Configuration args.
var globalConfig ConfSpec

func globalConfigInit() {
	// Read flags.
	flag.Var(&globalConfig.CommonsConfigs, "commons", "List of additional tmpl values files to alpply to main config.")
	flag.Var(&globalConfig.Packages, "p", "List of swarm packages names to process. Default - all.")
	flag.BoolVar(&globalConfig.Build, "build", false, "Build services images before deploy. Default: false")
	flag.BoolVar(&globalConfig.Debug, "debug", false, "Turn on debug logging. Default: false")
	flag.BoolVar(&globalConfig.DryRun, "dry-run", false, "Diffs output without any action. Default: false")
	flag.StringVar(&globalConfig.MainConfig, "f", "", "YAML manifest filename.")

	// Configuration args.
	flag.Parse()

	// Set unconfigured values.
	workdir, err := os.Getwd()

	checkErr(err)
	globalConfig.WorkDir = workdir
	globalConfig.DotDir = filepath.Join(".hiver")
	log.Debug("Working dir: ", globalConfig.WorkDir)
	globalConfig.SwarmPkgTmplFile = "main.tmpl"

}
