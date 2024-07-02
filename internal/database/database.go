package database

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Groskilled/Chirpy/internal/entities"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]entities.ChirpInterface `json:"chirps"`
	Users  map[int]entities.UserInterface  `json:"users"`
}

func NewDB(path string) (*DB, error) {
	newDB := DB{
		path: path,
	}
	err := newDB.EnsureDB()
	if err != nil {
		fmt.Println("Couldnt create connection to DB")
		return nil, err
	}
	return &newDB, nil
}

func (db *DB) EnsureDB() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
		// File does not exist, create it
		file, err := os.Create(db.path)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			return err
		}
		defer file.Close()

		// Optionally, you can write an empty JSON object to the file
		_, err = file.WriteString("{}")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return err
		}

		fmt.Printf("File %s created successfully.\n", db.path)
	} else if err != nil {
		fmt.Printf("Error checking file: %v\n", err)
	} else {
		fmt.Printf("File %s already exists.\n", db.path)
	}
	return nil
}

func (db *DB) WriteDB(dbStructure DBStructure) error {
	// Open the file for writing, create if it doesn't exist, truncate if it does
	fmt.Println("Write to db")
	file, err := os.Create(db.path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Encode the dbStructure into JSON format and write to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Optional: for pretty-printing

	if err := encoder.Encode(&dbStructure); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func (db *DB) LoadDB() (DBStructure, error) {
	var dbStructure DBStructure

	file, err := os.Open(db.path)
	if err != nil {
		return dbStructure, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return dbStructure, fmt.Errorf("error reading file: %v", err)
	}

	if err := json.Unmarshal(bytes, &dbStructure); err != nil {
		return dbStructure, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return dbStructure, nil
}
