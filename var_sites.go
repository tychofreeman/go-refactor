package refactor

import (
	"reflect"
	"go/token"
	"container/vector"
	"fmt"
)

type VarSites struct {
	varSites map[string]*vector.Vector
}

func NewVarSites() *VarSites {
	varSites := new(VarSites)
	varSites.varSites = make(map[string]*vector.Vector)
	return varSites
}

func (vs *VarSites) AddSite(name string, position token.Position) {
	if vs.varSites == nil {
		panic("VarSites.varSites was never set!!")
	}
	if vs.varSites[name] == nil {
		vs.varSites[name] = new(vector.Vector)
	}
	vs.varSites[name].Push(position)
}

func (vs *VarSites) GetSites(name string) (siteArray []token.Position) {
	siteVector := vs.varSites[name]
	if siteVector != nil {
		vectorLen := siteVector.Len()
		siteArray = make([]token.Position, vectorLen)
		for i := 0; i < vectorLen; i++ {
			vectorItem := siteVector.At(i)
			switch site := vectorItem.(type) {
				case token.Position:
					siteArray[i] = site
				default:
					panic(fmt.Sprintf("Found wrong type in siteVector at index %v - %v", i, reflect.Typeof(site)))
			}
		}
	}
	return 
}
