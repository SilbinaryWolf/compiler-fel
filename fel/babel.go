package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
)

func babelTranspileJavascript(code string) (string, error) {
	path, err := exec.LookPath("babel")
	if err != nil {
		return "", err
	}
	cmd := exec.Command(path, "--source-maps", "inline")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, code)
	}()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	compilationOutput, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	errorMessages, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", err
	}
	if len(errorMessages) > 0 {
		return "", errors.New(string(errorMessages))
	}
	if err := cmd.Wait(); err != nil {
		return "", err
	}
	return string(compilationOutput), nil
}
