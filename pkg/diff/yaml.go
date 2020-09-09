package diff

import (
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/pretty"
	"gopkg.in/yaml.v2"
)

// Yamls compare two yamls anf show diff.
// colored - turn on/off colored diff data.
func Yamls(yamlA []byte, yamlB []byte, colored bool) (string, error) {
	GreenColor := "%s"
	RedColor := "%s"
	if colored {
		GreenColor = "\033[1;32m%s\033[0m"
		RedColor = "\033[1;31m%s\033[0m"
	}
	// Unmarshal data.
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

	// Compare, join result to string and add colors.
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
