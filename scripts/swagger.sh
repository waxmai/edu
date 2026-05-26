#!/bin/bash
set -e

printf "\nRegenerating swagger doc\n\n"
time go run github.com/swaggo/swag/cmd/swag@v1.16.2 init
printf "\nDone.\n\n"
