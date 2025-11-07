package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}
}

func InitDB() (*pgx.Conn, error) {
	// Get connection details from environment variables
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// Create the connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	db, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}
	// Test the connection
	if err := db.Ping(context.Background()); err != nil {
		db.Close(context.Background())
		return nil, err
	}
	fmt.Println("Connected to PostgreSQL database!")
	return db, nil
}

func createAccount() {

}
