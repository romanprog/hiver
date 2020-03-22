package main

import (
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/yaml.v2"
)

func diffYamls(yamlA []byte, yamlB []byte, colored bool) (string, error) {
	GreenColor := "%s"
	RedColor := "%s"
	if colored {
		GreenColor = "\033[1;32m%s\033[0m"
		RedColor = "\033[1;31m%s\033[0m"
	}
	var yamlParsedA, yamlParsedB interface{}
	err := yaml.Unmarshal(yamlA, &yamlParsedA)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(yamlB, &yamlParsedB)
	if err != nil {
		return "", err
	}
	diffs := make([]string, 0)
	for _, s := range strings.Split(pretty.Compare(yamlParsedA, yamlParsedB), "\n") {
		switch {
		case strings.HasPrefix(s, "+"):
			diffs = append(diffs, fmt.Sprintf(GreenColor, s))
		case strings.HasPrefix(s, "-"):
			diffs = append(diffs, fmt.Sprintf(RedColor, s))
		}
	}
	return strings.Join(diffs, "\n"), nil
}
