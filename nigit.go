package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
)

func execProgram(program string, env []string, input string) bytes.Buffer {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(program)
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = env
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	if err != nil {
		fmt.Errorf("Execution of program %s failed\n", program)
		fmt.Errorf("%s\n", stderr)
	}

	return stdout
}

func urlPath(programPath string) string {
	return "/" + strings.TrimSuffix(filepath.Base(programPath), filepath.Ext(programPath))
}

func handleForm(programPath string, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(5 * 1000 * 1000)

	env := os.Environ()
	for k, v := range r.Form {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), strings.Join(v, " ")))
	}

	accept := r.Header.Get("Accept")
	env = append(env, fmt.Sprintf("%s=%s", "ACCEPT", accept))

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	stdout := execProgram(programPath, env, buf.String())

	// We reply with the requested content type as we do not know
	// what the program or script will ever return while the client does
	mediatype, _, err := mime.ParseMediaType(accept)
	if err == nil && mediatype != "*/*" {
		w.Header().Set("Content-Type", mediatype)
		stdout.WriteTo(w)
	} else {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, stdout.String())
	}
}

func serve(programPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			w.Header().Set("Content-Type", "application/json")
		default:
			handleForm(programPath, w, r)
		}
	})
}

func logRequests(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
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
		if !c.Args().Present() {
			fmt.Println("Please provide the names of the scripts to run under nigit")
			os.Exit(1)
		}
		fmt.Println("Serve from port " + c.GlobalString("port"))

		for _, program := range c.Args() {
			programPath, err := filepath.Abs(program)
			if err != nil {
				fmt.Printf("Cannot get path of %s\n", program)
				os.Exit(2)
			}

			programPath, err = exec.LookPath(programPath)
			if err != nil {
				fmt.Printf("Executable program %s not found\n", program)
				os.Exit(3)
			}

			fmt.Printf("Handle %s -> %s\n", urlPath(programPath), program)
			http.Handle(urlPath(programPath), serve(programPath))
		}
		http.ListenAndServe(":"+c.GlobalString("port"), logRequests(http.DefaultServeMux))
	}

	app.Run(os.Args)
}
