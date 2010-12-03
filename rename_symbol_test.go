package refactor

import (
	"testing"
)


func CallTwoLineGetVariableNameAt(row, col int) string {
	src := RefactorSource(TWO_LINE_DECL)

	name, _ := src.GetVariableNameAt(row, col)
	return name
}

func TestFindsNameOfVariableDeclStartingAtRowColumn(t *testing.T) {
	varName := CallTwoLineGetVariableNameAt(TWO_LINE_VAR_DECL_START_LINE, VAR_DECL_START_POS)

	if varName != TARGET_VAR_NAME {
		t.Fail()
	}
}

func TestFindsNameOfVariableDeclContainingRowColumn(t *testing.T) {
	varName := CallTwoLineGetVariableNameAt(TWO_LINE_VAR_DECL_START_LINE, VAR_DECL_START_POS);

	if varName != TARGET_VAR_NAME {
		t.Fail()
	}
}

func TestDoesntFindNameOfVariableDeclAfterColumn(t *testing.T) {
	varName := CallTwoLineGetVariableNameAt(TWO_LINE_VAR_DECL_START_LINE, VAR_DECL_START_POS - 2)
	if varName != "" {
		t.Fail()
	}
}

func TestDoesntFindNameOfVariableDeclOnNextRow(t *testing.T) {
	varName := CallTwoLineGetVariableNameAt(TWO_LINE_VAR_DECL_START_LINE-1, VAR_DECL_START_POS)
	if varName != "otherString" {
		t.Errorf("Expected [%v], got [%v]", varName, "")
	}
}

