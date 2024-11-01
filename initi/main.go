package initi

import (
	"fmt"

	"github.com/joho/godotenv"
)

func InitProj() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error at loading env: %v", err)
	}
}
