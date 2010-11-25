include $(GOROOT)/src/Make.inc

TARG=refactor
GOFILES=\
	rename_symbol.go\
	var_sites.go\
	refactor_visitor.go


all:$(TARG).a $(TARG).$O
	$(LD) -o $(TARG) $(TARG).$O

%.$O: %.go
	$(GC) $?

$(TARG).a: _obj/$(TARG).a
	cp $? ./

include $(GOROOT)/src/Make.pkg
