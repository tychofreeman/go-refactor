package refactor

import (
	"testing"
	"fmt"
	"go/token"
)

func CallScopedGetVariableNameAt(row, col int) string {
	src := RefactorSource(SIMPLE_DECL)

	return src.GetVariableNameAt(row, col)
}

const (
	FUNC_A = "func A() {\nvar integerVariable = 1\nintegerVariable += 1\n}\n"
	FUNC_B = "func B() {\nvar integerVariable = 1\nintegerVariable += 1\n}\n"
	FUNC_C = "func C() {\nvar integerVariable = 1\ngo func(){\nintegerVariable = 2\n}()\nintegerVariable++\n}\n"
)

func TestIdentifiesOnlyInScopeUsages(t *testing.T) {

	code := fmt.Sprintf("%v\n%v\n", FUNC_A, FUNC_B)
	src := RefactorSource(code)
	actual := src.PositionsForSymbolAt(3, 3)

	declPosition := token.Position{"", 15, 7, 5}
	usePosition  := token.Position{"", 0, 8, 1}
	expected := []token.Position {declPosition, usePosition}

	if assertHasAnyPositions(expected, actual) == true {
		t.Fail()
	}

	if len(actual) != len(expected) {
		t.Errorf("Returned wrong number of positions!")
	}

}

func TestIdentifiesAllScopesUnderDeclaration(t *testing.T) {

	code := fmt.Sprintf("%v\n%v\n%v\n", FUNC_A, FUNC_B, FUNC_C)
	src := RefactorSource(code)
	actual := src.PositionsForSymbolAt(15, 1)

	funcADeclPosition := token.Position{"", 0, 2, 5}
	funcAUsePosition := token.Position{"", 0, 3, 1}
	funcBDeclPosition := token.Position{"", 0, 7, 5}
	funcBUsePosition  := token.Position{"", 0, 8, 1}
	expected := []token.Position {funcADeclPosition, funcAUsePosition, funcBDeclPosition, funcBUsePosition}

	if len(actual) != 2 {
		t.Errorf("Returned zero actual positions!")
	}

	if assertHasAnyPositions(expected, actual) {
		t.Fail()
	}

	if !assertHasAnyPositions([]token.Position{token.Position{"", 0, 12, 5}}, actual) {
		t.Fail()
	}
}


func assertHasAnyPositions(expected []token.Position, actual chan token.Position) bool {
	found := false
	for _, ePos := range expected {
		for aPos := range actual {
			if ePos.Line == aPos.Line && ePos.Column == aPos.Column {
				found = true
			}
		}
	}
	return found
}
