package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
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
	app := cli.NewApp()
	app.Name = "nigit"
	app.Version = "0.1-alpha"
	app.Usage = "Expose any Program as HTTP API"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "port",
			Value:  "8000",
			Usage:  "HTTP port",
			EnvVar: "PORT",
		},
	}

	app.Action = func(c *cli.Context) {
		program := c.Args().First()
		if program == "" {
			fmt.Println("Please provide name of the script to run under nigit")
			os.Exit(1)
		}

		program, err := filepath.Abs(program)
		if err != nil {
			fmt.Printf("Cannot get path of %s\n", program)
			os.Exit(2)
		}

		programPath, err := exec.LookPath(program)
		if err != nil {
			fmt.Printf("Executable program %s not found\n", program)
			os.Exit(3)
		}

		serve := func(w http.ResponseWriter, r *http.Request) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)

			stdout := execProgram(programPath, buf.String())
			io.WriteString(w, stdout)
		}

		http.HandleFunc("/", serve)
		http.ListenAndServe(":"+c.GlobalString("port"), nil)
	}

	app.Run(os.Args)
}
