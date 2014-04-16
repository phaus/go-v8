package main

import "fmt"
import "../../"

const (
	code = `
        check([1,2,3,4,5,6,7,8,9,10])

        var a = get1()

        check(a)

        var b = get2()

        check(b[0])
        check(b[1])
    `
)

func get1() []int {
	s := make([]int, 10)
	for i := 0; i < len(s); i++ {
		s[i] = (i + 1) * 10
	}
	return s
}

func get2() ([]int, []int) {
	s1 := make([]int, 10)
	for i := 0; i < len(s1); i++ {
		s1[i] = (i + 1) * 10
	}

	s2 := make([]int, 10)
	for i := 0; i < len(s2); i++ {
		s2[i] = (i + 1) * 100
	}

	return s1, s2
}

func check(a []int) {
	fmt.Println(a)
}

func main() {
	engine := v8.NewEngine()
	script := engine.Compile([]byte(code), nil)

	if nil == script {
		panic("cannot compile")
	}

	global := engine.NewObjectTemplate()
	global.Bind("check", check)
	global.Bind("get1", get1)
	global.Bind("get2", get2)

	context := engine.NewContext(global)
	context.Scope(func(cs v8.ContextScope) {
		cs.Run(script)
	})
}
