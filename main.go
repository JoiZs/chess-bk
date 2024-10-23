package main

import (
	"fmt"
	"os"

	"github.com/JoiZs/chess-bk/initializer"
)

func main() {
	fmt.Println("Hello World")

	initializer.Init()

	fmt.Println(os.Getenv("test"))
}
