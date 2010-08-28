package refactor

import (
	"testing"
)

func CallUsageGetVariableNameAt(row, col int) string {
	src := RefactorSource(DECL_AND_USE)

	return src.GetVariableNameAt(row, col)
}

func TestFindsNameOfVariableUseAtStartOfUse(t *testing.T) {
	varName := CallUsageGetVariableNameAt(TWO_LINE_VAR_DECL_START_LINE, VAR_USE_START_POS)

	if varName != TARGET_VAR_NAME {
		t.Errorf("Expected [%v], but got [%v]", TARGET_VAR_NAME, varName)
	}
}
