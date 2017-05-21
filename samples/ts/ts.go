package main

import (
	"fmt"
	"io/ioutil"
	"os"

	v8 "github.com/saibing/go-v8"
)

func main() {
	code, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Read script file failed: %s\n", err)
		return
	}

	engine := v8.NewEngine()
	script := engine.Compile(code, nil)
	context := engine.NewContext(nil)

	context.Scope(func(cs v8.ContextScope) {
		result := cs.Run(script)
		println(result.ToString())
	})
}
