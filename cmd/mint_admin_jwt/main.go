package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/herodragmon/scalable-ecommerce/internal/auth"
)

func main() {
	_ = godotenv.Load() // 👈 REQUIRED HERE TOO

	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		panic("SECRET_KEY not set")
	}

	token, err := auth.MakeJWT(
		uuid.New(),
		"admin",
		secret,
		time.Hour*24,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}


