package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func execProgram(program string, input string) string {
	var stdout bytes.Buffer
	cmd := exec.Command(program)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	return stdout.String()
}

func main() {
	args := os.Args[1:]
	program, _ := filepath.Abs(args[0])

	serve := func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		stdout := execProgram(program, buf.String())
		io.WriteString(w, stdout)
	}

	http.HandleFunc("/", serve)
	http.ListenAndServe(":8000", nil)
}
