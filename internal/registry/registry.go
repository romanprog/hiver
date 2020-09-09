package registry

import (
	"fmt"

	"github.com/op/go-logging"
	"github.com/romanprog/hiver/internal/config"
	"github.com/romanprog/hiver/internal/runner"
)

var log = logging.MustGetLogger("hiver")

// Spec data for auth.
type Spec struct {
	URL  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

// Auth registry. Use docker login command.
func (r *Spec) Auth() error {
	return runner.BashExec(fmt.Sprintf("docker login %s -u %s -p %s", r.URL, r.User, r.Pass), r.Pass)
}

// AuthList - auth list of registries.
func AuthList(registries []Spec) {
	log.Infof("Processing auth registries list.")
	if len(registries) == 0 {
		log.Debug("Registries list is empty.")
		return
	}
	for _, r := range registries {
		log.Infof("Login to registry: %s", r.URL)
		if config.Global.DryRun {
			log.Noticef("Dry run. docker login %s -u %s -p ***", r.URL, r.User)
			continue
		}
		r.Auth()
	}
}
