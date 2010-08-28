package refactor

import (
	"testing"
)

const (
	SIMPLE_DECL = "var myString = \"Hello World\";"
	TWO_LINE_DECL = "var otherString = \"other string\";\nvar myString = \"Hello World\";"
	TARGET_VAR_NAME = "myString"
	VAR_DECL_START_LINE = 2
	VAR_DECL_START_POS = 5
)

func TestFindsNameOfVariableDeclStartingAtPosition(t *testing.T) {
	varName := CallGetVariableNameAt(1, VAR_DECL_START_POS)

	if varName != TARGET_VAR_NAME {
		t.Fail()
	}
}

func TestFindsNameOfVariableDeclContainingPosition(t *testing.T) {
	varName := CallGetVariableNameAt(1, VAR_DECL_START_POS);

	if varName != TARGET_VAR_NAME {
		t.Fail()
	}
}

func TestDoesntFindNameOfVariableDeclAfterPosition(t *testing.T) {
	varName := CallGetVariableNameAt(1, VAR_DECL_START_POS - 2)
	if varName != "" {
		t.Fail()
	}
}
