package main

import "refactor"
import "testing"

var tests = []testing.Test {
	testing.Test{ "refactor.TestIdentifiesOnlyInScopeUsagesInFuncA", refactor.TestIdentifiesOnlyInScopeUsagesInFuncA },
	testing.Test{ "refactor.TestIdentifiesOnlyInScopeUsagesInFuncB", refactor.TestIdentifiesOnlyInScopeUsagesInFuncB },
	testing.Test{ "refactor.TestIdentifiesAllScopesUnderDeclaration", refactor.TestIdentifiesAllScopesUnderDeclaration },
	testing.Test{ "refactor.TestIdentifierContainsPositionAtEndOfIdentifier", refactor.TestIdentifierContainsPositionAtEndOfIdentifier },
	testing.Test{ "refactor.TestIdentifierContainsPositionAtStartOfIdentifier", refactor.TestIdentifierContainsPositionAtStartOfIdentifier },
	testing.Test{ "refactor.TestFindsNameOfVariableDeclStartingAtPosition", refactor.TestFindsNameOfVariableDeclStartingAtPosition },
	testing.Test{ "refactor.TestFindsNameOfVariableDeclContainingPosition", refactor.TestFindsNameOfVariableDeclContainingPosition },
	testing.Test{ "refactor.TestDoesntFindNameOfVariableDeclAfterPosition", refactor.TestDoesntFindNameOfVariableDeclAfterPosition },
	testing.Test{ "refactor.TestFindsNameOfVariableDeclStartingAtRowColumn", refactor.TestFindsNameOfVariableDeclStartingAtRowColumn },
	testing.Test{ "refactor.TestFindsNameOfVariableDeclContainingRowColumn", refactor.TestFindsNameOfVariableDeclContainingRowColumn },
	testing.Test{ "refactor.TestDoesntFindNameOfVariableDeclAfterColumn", refactor.TestDoesntFindNameOfVariableDeclAfterColumn },
	testing.Test{ "refactor.TestDoesntFindNameOfVariableDeclOnNextRow", refactor.TestDoesntFindNameOfVariableDeclOnNextRow },
	testing.Test{ "refactor.TestFindsNameOfVariableUseAtStartOfUse", refactor.TestFindsNameOfVariableUseAtStartOfUse },
	testing.Test{ "refactor.TestFindsNameOfVariableUseIfNameContainsPosition", refactor.TestFindsNameOfVariableUseIfNameContainsPosition },
}
var benchmarks = []testing.Benchmark {
}

func main() {
	testing.Main(tests);
	testing.RunBenchmarks(benchmarks)
}
