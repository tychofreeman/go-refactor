package refactor

import (
	"reflect"
	"go/token"
	"container/vector"
	"fmt"
)

type Scope struct {
	parent *Scope
	children []*Scope
	childCount int
	positions map[string]*vector.Vector
	decls map[string]token.Position
}

func NewChildScope(parent *Scope) (child *Scope) {
	child = NewScope()
	child.parent = parent
	parent.childCount++
	parent.children[parent.childCount] = child
	return
}

func NewScope() (scope *Scope) {
	scope = new(Scope)
	scope.positions = make(map[string]*vector.Vector)
	scope.decls = make(map[string]token.Position)
	scope.children = make([]*Scope, 5)
	return scope
}

func (scope *Scope) AddDefn(name string, position token.Position) {
	scope.decls[name] = position
}

func (scope *Scope) HasDefn(name string) bool {
	_, ok := scope.decls[name]
	return ok
}

func (scope *Scope) GetDefn(name string) token.Position {
	return scope.decls[name]
}

func (scope *Scope) AddSite(name string, position token.Position) {
	if scope.positions == nil {
		panic("Scope.positions was never set!!")
	}
	if scope.positions[name] == nil {
		scope.positions[name] = new(vector.Vector)
	}
	scope.positions[name].Push(position)
}

func (scope *Scope) GetSites(name string) (siteArray []token.Position) {
	siteVector := scope.positions[name]
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
