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
	"time"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("nigit")
var colorFormat = logging.MustStringFormatter(
	`%{color}%{level:.7s} ▶ %{message}%{color:reset}`,
)
var uncoloredFormat = logging.MustStringFormatter(
	`%{level:.7s} ▶ %{message}`,
)

func execProgram(program string, extraEnv []string, input string, timeout int) bytes.Buffer {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	env := append(os.Environ(), extraEnv...)

	programName := filepath.Base(program)
	cmd := exec.Command(program)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = env

	reportFailure := func() {
		log.Errorf(
			"Execution of program %s failed with %s\n%s",
			programName,
			cmd.ProcessState.String(),
			strings.Trim(stderr.String(), "\n"))
	}

	if err := cmd.Start(); err != nil {
		reportFailure()
		return stdout
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("Cannot kill process: ", err)
		}
		log.Debugf("Process %s killed", programName)
	case err := <-done:
		if err != nil {
			reportFailure()
		} else {
			log.Debugf("Executed %s without error in %s", programName, cmd.ProcessState.UserTime())
		}
	}

	return stdout
}

func urlPath(programPath string) string {
	return "/" + strings.TrimSuffix(filepath.Base(programPath), filepath.Ext(programPath))
}

func handleForm(programPath string, w http.ResponseWriter, r *http.Request, timeout int) {
	r.ParseMultipartForm(5 * 1000 * 1000)

	// All form arguments are injected into the environment of the executed child program
	var env []string
	for k, v := range r.Form {
		env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), strings.Join(v, " ")))
	}

	// Important HTTP headers are passed to the child program so it can decide what content it wants to output
	accept := r.Header.Get("Accept")
	env = append(env, fmt.Sprintf("%s=%s", "ACCEPT", accept))

	env = append(env, fmt.Sprintf("%s=%s", "HOST", r.Header.Get("Host")))
	env = append(env, fmt.Sprintf("%s=%s", "USER_AGENT", r.Header.Get("User-Agent")))

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	stdout := execProgram(programPath, env, buf.String(), timeout)

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

func serve(programPath string, timeout int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			w.Header().Set("Content-Type", "application/json")
		default:
			handleForm(programPath, w, r, timeout)
		}
	})
}

func logRequests(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s", r.Method, r.URL)
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
		cli.IntFlag{
			Name:   "timeout",
			Value:  5,
			Usage:  "Timeout in seconds after process is stopped",
			EnvVar: "TIMEOUT",
		},
		cli.BoolFlag{
			Name:   "no-color",
			Usage:  "Do not colorize output",
			EnvVar: "NO_COLOR",
		},
	}

	app.Action = func(c *cli.Context) {
		if !c.Args().Present() {
			fmt.Println("Please provide the names of the scripts to run under nigit")
			os.Exit(1)
		}

		if c.GlobalBool("no-color") {
			logging.SetFormatter(uncoloredFormat)
		} else {
			logging.SetFormatter(colorFormat)
		}

		log.Infof("Serve from port %s with %d seconds timeout", c.GlobalString("port"), c.GlobalInt("timeout"))

		for _, program := range c.Args() {
			programPath, err := filepath.Abs(program)
			if err != nil {
				log.Fatalf("Cannot get path of %s", program)
			}

			programPath, err = exec.LookPath(programPath)
			if err != nil {
				log.Fatalf("Executable program %s not found", program)
			}

			log.Infof("Handle %s -> %s", urlPath(programPath), program)
			http.Handle(urlPath(programPath), serve(programPath, c.GlobalInt("timeout")))
		}
		http.ListenAndServe(":"+c.GlobalString("port"), logRequests(http.DefaultServeMux))
	}

	app.Run(os.Args)
}
