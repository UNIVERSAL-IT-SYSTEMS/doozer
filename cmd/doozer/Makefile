include ../../Make.inc

VERSION ?= $(shell ../../ver.sh)

TARG=doozer
GOFILES=\
	add.go\
	del.go\
	doozer.go\
	get.go\
	help.go\
	nop.go\
	rev.go\
	set.go\
	stat.go\
	vers.go\
	wait.go\
	watch.go\
	find.go\
	touch.go\
	watchService.go\
	addService.go\
	delService.go\

include $(GOROOT)/src/Make.cmd

vers.go:
	printf 'package main\n\nconst version = `%s`\n' '$(VERSION)' > $@
