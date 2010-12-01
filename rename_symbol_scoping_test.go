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

func TestIdentifiesOnlyInScopeUsagesInFuncA(t *testing.T) {

	code := fmt.Sprintf("%v\n%v\n", FUNC_A, FUNC_B)
	src := RefactorSource(code)
	src.PositionsForSymbolAt(3, 3)
	actual := copyChannelToArray(src.gimme)

	decl1Position := token.Position{"", 0, 7, 5}
	use1Position := token.Position{"", 0, 8, 1}
	decl2Position := token.Position{"", 0, 11, 5}
	use2Position := token.Position{"", 0, 13, 1}
	use3Position := token.Position{"", 0, 15, 1}
	unexpected := []token.Position {decl1Position, use1Position, decl2Position, use2Position, use3Position}
	expected := []token.Position { token.Position{"", 0, 3, 1}, token.Position{"", 0, 2, 5} }

	if assertHasAnyPositions(unexpected, actual) == true {
		t.Errorf("Contains unexpected positions: %v\n", unexpected)
	}

	if assertHasAnyPositions(expected, actual) == false {
		t.Errorf("Returned positions %v instead of %v!", actual, expected)
	}
}

func TestIdentifiesOnlyInScopeUsagesInFuncB(t *testing.T) {

	code := fmt.Sprintf("%v\n%v\n", FUNC_A, FUNC_B)
	src := RefactorSource(code)
	src.PositionsForSymbolAt(7, 6)
	actual := copyChannelToArray(src.gimme)

	decl1Position := token.Position{"", 0, 2, 5}
	use1Position := token.Position{"", 0, 3, 1}
	decl2Position := token.Position{"", 0, 10, 5}
	use2Position := token.Position{"", 0, 12, 1}
	use3Position := token.Position{"", 0, 14, 1}
	unexpected := []token.Position {decl1Position, use1Position, decl2Position, use2Position, use3Position}
	expected := []token.Position { token.Position{"", 0, 8, 1}, token.Position{"", 0, 7, 5} }

	if assertHasAnyPositions(unexpected, actual) == true {
		t.Errorf("Contains unexpected positions: %v\n", unexpected)
	}

	if assertHasAnyPositions(expected, actual) == false {
		t.Errorf("Returned positions %v instead of %v!", actual, expected)
	}
}

func TestIdentifiesAllScopesUnderDeclaration(t *testing.T) {

	code := fmt.Sprintf("%v\n%v\n%v\n", FUNC_A, FUNC_B, FUNC_C)
	src := RefactorSource(code)
	src.PositionsForSymbolAt(16, 1)
	actual := copyChannelToArray(src.gimme)

	funcADeclPosition := token.Position{"", 0, 2, 5}
	funcAUsePosition := token.Position{"", 0, 3, 1}
	funcBDeclPosition := token.Position{"", 0, 7, 5}
	funcBUsePosition  := token.Position{"", 0, 8, 1}
	expected := []token.Position {funcADeclPosition, funcAUsePosition, funcBDeclPosition, funcBUsePosition}

	if assertHasAnyPositions(expected, actual) {
		t.Errorf("Has wrong positions: %v vs %v", expected, actual)
	}


	if !assertHasAnyPositions([]token.Position{token.Position{"", 0, 12, 5}}, actual) {
		t.Errorf("Missing position 12:5 (has %v)", actual)
	}
}

func TestIdentifierContainsPositionAtEndOfIdentifier(t *testing.T) {
	if !identContainsPosition("abcdefg", token.Position{"", 0, 12, 5}, 12,  5 - 1 + len("abcdefg")) {
		t.Errorf("Did not identify position at end of identifier")
	}
}

func TestIdentifierContainsPositionAtStartOfIdentifier(t *testing.T) {
	if !identContainsPosition("abcdefg", token.Position{"", 0, 12, 5}, 12,  5) {
		t.Errorf("Did not identify position at start of identifier")
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
