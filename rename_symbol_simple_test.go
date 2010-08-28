package refactor

import (
	"testing"
)

const (
	SIMPLE_DECL = "var myString = \"Hello World\";"
	TWO_LINE_DECL = "var otherString = \"other string\";\nvar myString = \"Hello World\";"
	DECL_AND_USE = "var myString = \"Hello World\"\nvar a = myString;";
	TARGET_VAR_NAME = "myString"
	TWO_LINE_VAR_DECL_START_LINE = 2
	VAR_USE_START_POS = 9
	VAR_DECL_START_POS = 5
)

func CallSimpleGetVariableNameAt(row, col int) string {
	src := RefactorSource(SIMPLE_DECL)

	return src.GetVariableNameAt(row, col)
}

func TestFindsNameOfVariableDeclStartingAtPosition(t *testing.T) {
	varName := CallSimpleGetVariableNameAt(1, VAR_DECL_START_POS)

	if varName != TARGET_VAR_NAME {
		t.Fail()
	}
}

func TestFindsNameOfVariableDeclContainingPosition(t *testing.T) {
	varName := CallSimpleGetVariableNameAt(1, VAR_DECL_START_POS);

	if varName != TARGET_VAR_NAME {
		t.Fail()
	}
}

func TestDoesntFindNameOfVariableDeclAfterPosition(t *testing.T) {
	varName := CallSimpleGetVariableNameAt(1, VAR_DECL_START_POS - 2)
	if varName != "" {
		t.Fail()
	}
}
