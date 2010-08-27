include $(GOROOT)/src/Make.$(GOARCH)

TARG=refactor
GOFILES=\
	rename_symbol.go


include $(GOROOT)/src/Make.pkg
