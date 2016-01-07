# nigit [![Build Status](https://travis-ci.org/lukasmartinelli/nigit.svg)](https://travis-ci.org/lukasmartinelli/nigit) ![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)

<img align="right" alt="nigit cat logo" src="nigit.png" />

Expose a program with a simple call to `nigit <script>` to the web.
The small web server wraps around the program and exposes it as HTTP API.
This comes in handy whenever you want to expose a legacy
program to the internet without writing a web application and doing complicated
subprocessing yourself.

## Get Started

In this example we create a service to download the PDF version of websites using the
[wkhtmltopdf](http://wkhtmltopdf.org/) tool.

1. Create the bash script `html2pdf.sh` and make it executable with `chmod +x html2pdf.sh`.
  ```bash
  #!/bin/bash
  wkhtmltopdf "$URL" page.pdf > /dev/null 2>&1
  cat page.pdf
  ```

2. Start up the server with `nigit html2pdf.sh`.
3. Download the PDF with `  curl -o google.pdf http://localhost:8000/html2pdf?url=http://google.com`

And that's all you needed to do in order to expose `wkhtml2pdf` as useful webservice.

## Install

You can download a single binary for Linux, OSX or Windows.

**OSX**

```bash
wget -O nigit https://github.com/lukasmartinelli/nigit/releases/download/v0.2/nigit_darwin_amd64
chmod +x nigit

./nigit --help
```

**Linux**

```bash
wget -O nigit https://github.com/lukasmartinelli/nigit/releases/download/v0.2/nigit_linux_amd64
chmod +x nigit

./nigit --help
```

**Install from Source**

```bash
go get github.com/lukasmartinelli/nigit
```

If you are using Windows or 32-bit architectures you need to [download the appropriate binary
yourself](https://github.com/lukasmartinelli/nigit/releases/latest).

## Examples

How you can use `nigit` to build small and concise services:

- PDF build service using `pdflatex`
- Convert DOCX files to Markdown with `pandoc`
- Image cropping with `imagemagick`
- Convert WAV to MP4 with `avconf`
- Transpile code with `BabelJS`
- Lint Shell scripts with `shellcheck`

I use `nigit` to create a HTTP API to programming language linters
in my [lintfox project](https://github.com/lukasmartinelli/lintfox).

## Usage

### Pass Arguments

Form arguments or query strings are passed as environment variables into the program.

```bash
#!/bin/bash
echo "$MY_ARGUMENT"
```

You can specify them as form values or alternatively post a JSON file which is more convenient
in JavaScript.

```bash
# pass as form values
curl -X POST -F my_argument=test http://localhost:8000/
# pass as query string
curl http://localhost:8000/?my_argument=test
# pass as JSON
curl -X POST -H "Content-Type: application/json" \
  -d '{"envs": ["MY_ARGUMENT=test"]}' \
  http://localhost:8000/hlint
```

### Upload File

Uploaded content is provided as `stdin` to the file.
You can either specify it as file upload or form value.
In both cases the field must be named `stdin`.

```bash
# set value of stdin in form
curl -X POST -F stdin="Ping" http://localhost:8000/
# for short input you can even use query strings
curl http://localhost:8000/?stdin=Ping
# post a file to the web api
curl -X POST -F stdin=@greetings.txt http://localhost:8000/
```

You can also specify a `stdin` field in JSON to pass something to the script.

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"stdin": "Ping", "envs": ["MY_ARGUMENT=test"]}' \
  http://localhost:8000/hlint
```

If you are using a Bash script as wrapper you can read in the stdin with `cat`.

```bash
#!/bin/bash
greetings=$(cat)
echo "$greetings"
```

### Serve Multiple Files

`nigit` can also serve multiple scripts under different paths if you
append more programs as arguments.

```bash
nigit echo.sh curl.sh lint.sh
```

This will serve each script under a different HTTP route.

```bash
Handle /echo -> echo.sh
Handle /curl -> curl.sh
Handle /lint -> lint.sh
```

### Mime Type

`nigit` serves the response either as `text/plain` if no `Accept` header is specified or
exactly with the mime type specified by the `Accept` header.

If you wrap around a program that outputs valid JSON you need to set the `Accept` header and you are good.

```bash
curl -H "Accept: application/json" http://localhost:8000/
```

### Use together with Docker

`nigit` fits perfectly into the Docker ecosystem. You can install `nigit` into a Docker
container and wrap around a program that requires complex dependencies.

Create a `Dockerfile` with `nigit` and your dependencies for the shell script.
In this example we provide `shellcheck` as a service.

```dockerfile
FROM debian:jessie

RUN apt-get update \
 && apt-get install -y --no-install-recommends shellcheck \
 && rm -rf /var/lib/apt/lists/* \

# install nigit
RUN wget --quiet -O /usr/bin/nigit https://github.com/lukasmartinelli/nigit/releases/download/v0.2/nigit_linux_amd64 \
 && chmod +x /usr/bin/nigit

# copy shell scripts
COPY . /usr/src/app/
WORKDIR /usr/src/app

EXPOSE 8000
CMD ["nigit", "--timeout", "5", "shellcheck.sh"]
```

Now create a bash script to wrap around `shellcheck`.
We specify the `json` output formatter so that a web client could
consume the API.

```bash
#!/bin/bash

function clone_repo() {
    local working_dir=$(mktemp -dt "lint.XXX")
    local git_output=$(git clone --quiet "$GIT_REPOSITORY" "$working_dir")
    echo "$working_dir"
}

function find_files() {
    local path="$1"
    local extension="$2"
    find "$path" -type f -name "*$extension"
}

function lint() {
    local repo_path=$(clone_repo)
    shellcheck --format=json $(find_files "$repo_path" "sh") || suppress_lint_error
    trap "rm -rf $repo_path" EXIT
}

lint
```

And now you can send links to Git repositories to your service to check them for Bash errors.

```bash
curl -H "Accept: application/json" \
http://localhost:8000/shellcheck?git_repository=https://github.com/lukasmartinelli/nigit.git
```

## Develop

You need a [Go workspace](https://golang.org/doc/code.html) to get started. 

### Install Dependencies

Several dependencies are required.

```
go get "github.com/codegangsta/cli"
go get "github.com/op/go-logging"
```

### Build

Create a executable using the standard Go build tool.

```
go build
```

### Cross Compile Release

We use [gox](https://github.com/mitchellh/gox) to create distributable
binaries for Windows, OSX and Linux.

```bash
docker run --rm -v "$(pwd)":/usr/src/nigit -w /usr/src/nigit tcnksm/gox:1.4.2-light
```

## Security

It is quite dangerous to expose a shell script to the internet. I also haven't tested any exploits
yet but my guess is a shell script takes input from external is always vulnerable.
