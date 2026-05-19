package main

import (
	"fmt"
	"log"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/seeder"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	db := configs.InitDB()

	// Seed database
	if err := seeder.SeedDatabase(db); err != nil {
		fmt.Printf("Error seeding database: %v\n", err)
		log.Fatal(err)
	}

	fmt.Println("\n✅ Database seeding completed!")
}
