#!/bin/sh
set -x
if [ "$1" = gnu ]; then
	flags='-g -std=c89 -fsanitize=undefined,address -pedantic -fdiagnostics-column-unit=byte'
fi
c89 $flags source/main.c -o notlang
