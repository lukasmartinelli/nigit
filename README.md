# nigit [![Build Status](https://travis-ci.org/lukasmartinelli/nigit.svg)](https://travis-ci.org/lukasmartinelli/nigit)

A web server that wraps around programs and shell scripts and exposes them as API.

## Get Started

Create a bash script `echo.sh` which will echo the input from `stdin`.

```
#!/bin/bash
read input
echo "$input"
```

Now execute it with `nigit echo.sh`.
A HTTP server has now been started on `localhost:8000`.
Let's execute an API call.

```
curl -X POST -d "Can you hear me?" http://localhost:8000/
```

You should now receive the same content you sent to the server.

```
Can you hear me?
```

## Use Cases

This use case comes in handy everywhere where you want to expose a legacy
program to the internet to use it as a service without writing a wrapper
script in a different language.

- Generate PDF
- Compile C++ code
- Lint code
