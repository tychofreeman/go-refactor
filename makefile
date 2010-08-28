include $(GOROOT)/src/Make.$(GOARCH)

TARG=refactor
GOFILES=\
	rename_symbol.go\
	var_sites.go\
	refactor_visitor.go
#	symbol_table.go


include $(GOROOT)/src/Make.pkg
