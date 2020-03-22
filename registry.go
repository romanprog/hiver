package main

import (
	"fmt"
)

type RegisrySpec struct {
	Url  string `yaml:"url"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

func (r *RegisrySpec) Auth() {
	err := commandExec(fmt.Sprintf("docker login %s -u %s -p %s", r.Url, r.User, r.Pass), r.Pass)
	checkErr(err)
}

func authRegistries(spec *hiverSpec) {
	log.Infof("Processing auth registries list.")
	if len(spec.Registries) == 0 {
		log.Debug("Registries list is empty.")
		return
	}
	for _, r := range spec.Registries {
		log.Infof("Login to registry: %s", r.Url)
		if globalConfig.DryRun {
			log.Noticef("Dry run. docker login %s -u %s -p ***", r.Url, r.User)
			continue
		}
		r.Auth()
	}
}
