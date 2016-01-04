# nigit [![Build Status](https://travis-ci.org/lukasmartinelli/nigit.svg)](https://travis-ci.org/lukasmartinelli/nigit) ![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)

<img align="right" alt="nigit cat logo" src="logo.png" />

Expose any program with a simple call to `nigit <script>` to the web.
The small web server wraps around the program and exposes them as HTTP API.
This comes in handy everywhere where you want to expose a legacy
program to the internet to use it as a service without writing a wrapper script
in a different language.

## Get Started

Create a bash script `echo.sh` which will echo the input from `stdin`.

```bash
#!/bin/bash
read input
echo "$input"
```

Now execute it with `nigit echo.sh`.
A HTTP server has now been started on `localhost:8000`.
Let's execute an API call.

```bash
curl -X POST -d "Can you hear me?" http://localhost:8000/
```

You should now receive the same content you sent to the server.

```
Can you hear me?
```

## Install

You can download a single binary for Linux, OSX or Windows.

**OSX**

```bash
wget -O nigit https://github.com/lukasmartinelli/pipecat/releases/download/v0.1-alpha/nigit_darwin_amd64
chmod +x nigit

./nigit --help
```

**Linux**

```bash
wget -O nigit https://github.com/lukasmartinelli/pipecat/releases/download/v0.1-alpha/nigit_linux_amd64
chmod +x nigit

./nigit --help
```

**Install from Source**

```bash
go get github.com/lukasmartinelli/nigit
```

If you are using Windows or 32-bit architectures you need to [download the appropriate binary
yourself](https://github.com/lukasmartinelli/nigit/releases/latest).

## Advanced Usage

### Pass Arguments

Form pairs or other JSON values are passed as environment variables into the program.

```bash
#!/bin/bash
echo "$MY_ARGUMENT"
```

And now pass an argument.

### Serve Multiple Files

`nigit` can also serve multiple scripts under different paths if you
append more programs as arguments.

```bash
nigit echo.sh curl.sh lint.sh
```

If you pass form or JSON values they will be provided as uppercase
env vars in your program.

```bash
curl -X POST -F my_argument=test http://localhost:8000/
```

This will serve each script under a different HTTP route.

```bash
Serve from port 8000
Handle /echo -> echo.sh
Handle /curl -> curl.sh
Handle /lint -> lint.sh
```

## Use Cases

This use case comes in handy everywhere where you want to expose a legacy
program to the internet to use it as a service without writing a wrapper
script in a different language.

- Generate PDF
- Compile C++ code
- Lint code

## Cross Compile Release

We use [gox](https://github.com/mitchellh/gox) to create distributable
binaries for Windows, OSX and Linux.

```bash
docker run --rm -v "$(pwd)":/usr/src/nigit -w /usr/src/nigit tcnksm/gox:1.4.2-light
```
