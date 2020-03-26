package main

import (
	"github.com/markbates/pkger"
	"go.borchero.com/cuckoo/cmd"
)

func main() {
	pkger.Include("utils/templates/")
	cmd.Execute()
}
