package config

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("hiver")

// ConfSpec type for global config.
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
var Global ConfSpec

func init() {
	// Read flags.
	flag.Var(&Global.CommonsConfigs, "c", "List of additional tmpl values files to alpply to main config.")
	flag.Var(&Global.Packages, "p", "List of swarm packages names to process. Default - all.")
	flag.BoolVar(&Global.Build, "build", false, "Build services images before deploy. Default: false")
	flag.BoolVar(&Global.Debug, "debug", false, "Turn on debug logging. Default: false")
	flag.BoolVar(&Global.DryRun, "dry-run", false, "Diffs output without any action. Default: false")
	flag.StringVar(&Global.MainConfig, "f", "hiver.yaml", "YAML manifest filename.")

	// Configuration args.
	flag.Parse()
	// Set values.
	workdir, err := os.Getwd()
	if err != nil {
		log.Fatalf(err.Error())
	}
	Global.WorkDir = workdir
	Global.DotDir = filepath.Join(".hiver")
	log.Debug("Working dir: ", Global.WorkDir)
	Global.SwarmPkgTmplFile = "main.yaml"

}
