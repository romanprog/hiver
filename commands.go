package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
)

func stringHideSecrets(str string, secrets ...string) string {
	hiddenStr := str
	for _, s := range secrets {
		hiddenStr = strings.Replace(hiddenStr, s, "***", -1)
	}
	return hiddenStr
}

func commandExecCommon(command string, outputBuff io.Writer, errBuff io.Writer, secrets ...string) error {
	hiddenCommand := stringHideSecrets(command, secrets...)

	cmd := exec.Command("bash", "-c", command)
	log.Debugf("Executing command \"%s\"", hiddenCommand)

	cmd.Stdout = outputBuff
	cmd.Stderr = errBuff
	err := cmd.Run()

	return err
}

func commandExecOutput(command string, secrets ...string) (string, string, error) {
	output := &bytes.Buffer{}
	runerr := &bytes.Buffer{}
	err := commandExecCommon(command, output, runerr, secrets...)
	return output.String(), runerr.String(), err
}

func commandExec(command string, secrets ...string) error {

	var output io.Writer

	if globalConfig.Debug {
		output = os.Stdin
	}
	runerr := &bytes.Buffer{}
	err := commandExecCommon(command, output, runerr, secrets...)
	if err != nil {
		log.Debugf("Command exited with error. Error output:\n%s", runerr.String())
		return err
	}
	return nil
}
