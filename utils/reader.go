package utils

import (
	"fmt"
	"os"
	"strings"
)

// validate and read file content
func ReadFile(fileName string) ([]string, error) {
	//read conteent
	contentBytes, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	//split by new line
	lines := strings.Split(string(contentBytes))
	//validate lines
	//stopped here
}
