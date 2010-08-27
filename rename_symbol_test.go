package refactor

import (
	"testing"
)

func TestFindsNameOfVariableAtPosition(t *testing.T) {
	var src = RefactorSource("var myString = \"Hello World\";")

	var varDecl = src.GetVariableNameAt(4);

	if varDecl != "myString" {
		t.Fail()
	}
}
