package refactor

import (
	"testing"
)

const (
	SIMPLE_DECL = "var myString = \"Hello World\";"
	VAR_NAME = "myString"
	VAR_DECL_START_POS = 5
)

func CallGetVariableNameAt(pos int) string {
	src := RefactorSource(SIMPLE_DECL)

	return src.GetVariableNameAt(pos);
}

func TestFindsNameOfVariableDeclStartingAtPosition(t *testing.T) {
	varName := CallGetVariableNameAt(VAR_DECL_START_POS)

	if varName != VAR_NAME {
		t.Fail()
	}
}

func TestFindsNameOfVariableDeclContainingPosition(t *testing.T) {
	varName := CallGetVariableNameAt(VAR_DECL_START_POS);

	if varName != VAR_NAME {
		t.Fail()
	}
}

func TestDoesntFindNameOfVariableDeclAfterPosition(t *testing.T) {
	varName := CallGetVariableNameAt(VAR_DECL_START_POS - 2)
	if varName != "" {
		t.Fail()
	}
}
