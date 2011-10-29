package main

import (
	"flag"
	"fmt"
	"./refactor"
)

func main() {
	var row int
	var col int
	var file string
	var action string
	var name string

	flag.IntVar(&row, "row", -1, "Row of the symbol")
	flag.IntVar(&col, "col", -1, "Col of the symbol")
	flag.StringVar(&file, "file", "", "Name of the file")
	flag.StringVar(&action, "action", "", "Action to perform")
	flag.StringVar(&name, "name", "", "Name to give")

	flag.Parse()

	fmt.Printf("File: %v Row: %v Col: %v\n", file, row, col)


	rs := refactor.RefactorFile(file)
	currName, scope := rs.GetVariableNameAt(row, col)
	fmt.Printf("Rename variable %v to %v\n", currName, name)
	for _, site := range scope.GetSites(currName) {
		fmt.Printf("\t--%v\n", site)
	}
}
