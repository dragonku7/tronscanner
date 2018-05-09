package main

import (
	// "fmt"
	"os"
	"tronscanner/scanner"
	"tronscanner/types"
)

func main() {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println(r)
	// 	}
	// }()
	exit := make(chan bool)
	//start data writer
	scan, err := scanner.NewScanner()
	if err != nil {
		os.Exit(types.ERROR_INIT_FAILED)
	}
	scan.Start()
	defer scan.Stop()
	<-exit
}
