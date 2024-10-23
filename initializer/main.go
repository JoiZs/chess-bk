package initializer

import (
	"github.com/joho/godotenv"
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		panic("Err at loading .env")
	}
}
