#!/bin/bash
wkhtmltopdf "$URL" page.pdf > /dev/null 2>&1
cat page.pdf
