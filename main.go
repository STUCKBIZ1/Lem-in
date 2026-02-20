package main

import (
	"fmt"
	"lem-in/utils"
	"os"
)

func main() {
	//validate number of arguments
	if len(os.Args) != 2 {
		fmt.Println("Error: only one argument is allowed")
	}
	//extract argument
	fileName := os.Args[1]
	//validate and read fileName
	utils.ReadFile(fileName)
}
