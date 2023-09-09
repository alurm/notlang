os = $(shell uname -o)
android = $(filter Android, $(os))
sanitizer = $(if $(android),,-fsanitize=undefined)
compiler = gcc -g$(sanitizer) -Wno-gnu-designator
sources = main.c
headers = main.h
program : $(sources) makefile $(headers)
	$(compiler) -o $@ $(sources)
