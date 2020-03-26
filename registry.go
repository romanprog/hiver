package main

import (
	"fmt"
)

// RegistrySpec data for auth.
type RegistrySpec struct {
	URL  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

// Auth registry. Use docker login command.
func (r *RegistrySpec) Auth() {
	err := commandExec(fmt.Sprintf("docker login %s -u %s -p %s", r.URL, r.User, r.Pass), r.Pass)
	checkErr(err)
}

// Auth list of registries.
func authRegistries(spec *hiverSpec) {
	log.Infof("Processing auth registries list.")
	if len(spec.Registries) == 0 {
		log.Debug("Registries list is empty.")
		return
	}
	for _, r := range spec.Registries {
		log.Infof("Login to registry: %s", r.URL)
		if globalConfig.DryRun {
			log.Noticef("Dry run. docker login %s -u %s -p ***", r.URL, r.User)
			continue
		}
		r.Auth()
	}
}
