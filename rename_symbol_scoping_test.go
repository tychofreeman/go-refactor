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
)

func TestIdentifiesOnlyInScopeUsages(t *testing.T) {

	code := fmt.Sprintf("%v\n%v\n", FUNC_A, FUNC_B)
	src := RefactorSource(code)
	actual := src.PositionsForSymbolAt(3, 3)

	declPosition := token.Position{"", 15, 7, 5}
	usePosition  := token.Position{"", 0, 8, 1}
	expected := []token.Position {declPosition, usePosition}

	if assertHasAnyPositions(expected, actual) {
		t.Fail()
	}
}


func assertHasAnyPositions(expected []token.Position, actual []token.Position) bool {
	found := false
	for _, ePos := range expected {
		for _, aPos := range actual {
			if ePos.Line == aPos.Line && ePos.Column == aPos.Column {
				found = true
			}
		}
	}
	return found
}
