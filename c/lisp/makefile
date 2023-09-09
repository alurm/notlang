compiler = gcc -g -std=c99 -fsanitize=undefined -fplan9-extensions -fdiagnostics-column-unit=byte
sources = main.c
headers = main.h
program : $(sources) makefile $(headers)
	$(compiler) -o $@ $(sources)
