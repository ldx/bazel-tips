package main

import (
	"fmt"

	"github.com/ldx/bazel_tips/pkg/mypackage"
)

func main() {
	mypackage.DoSomething()
	fmt.Println("Hello, Bazel!")
}
