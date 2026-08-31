package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"my-kv-store/internal/db"
)

func main() {
	fmt.Println("=== Starting Key-Value Store Engine ===")
	opts := db.DefaultOptions("./data")
	database, err := db.Open(opts)
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	fmt.Println("Database ready! Available commands:")
	fmt.Println("  SET <key> <value>")
	fmt.Println("  GET <key>")
	fmt.Println("  DELETE <key>")
	fmt.Println("  EXIT")
	fmt.Println("-------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("kv> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		command := strings.ToUpper(parts[0])

		switch command {
		case "EXIT", "QUIT":
			fmt.Println("Shutting down database...")
			return

		case "SET":
			if len(parts) < 3 {
				fmt.Println("Usage: SET <key> <value>")
				continue
			}
			key := []byte(parts[1])
			val := []byte(parts[2])
			if err := database.Put(key, val); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "GET":
			if len(parts) < 2 {
				fmt.Println("Usage: GET <key>")
				continue
			}
			key := []byte(parts[1])
			val, err := database.Get(key)
			if err != nil {
				fmt.Printf("(nil) - %v\n", err)
			} else {
				fmt.Printf("%s\n", string(val))
			}

		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("Usage: DELETE <key>")
				continue
			}
			key := []byte(parts[1])
			if err := database.Delete(key); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
	}
}