package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Id   int
	Body string
}

func (db *DB) ensureDB() error {
	filename := "database.json"

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// File does not exist, create it
		file, err := os.Create(filename)
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

		fmt.Printf("File %s created successfully.\n", filename)
	} else if err != nil {
		fmt.Printf("Error checking file: %v\n", err)
	} else {
		fmt.Printf("File %s already exists.\n", filename)
	}
	return nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	// Open the file for writing, create if it doesn't exist, truncate if it does
	file, err := os.Create("database.json")
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

func (db *DB) loadDB() (DBStructure, error) {
	var dbStructure DBStructure

	file, err := os.Open("database.json")
	if err != nil {
		return dbStructure, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return dbStructure, fmt.Errorf("error reading file: %v", err)
	}

	if err := json.Unmarshal(bytes, &dbStructure); err != nil {
		return dbStructure, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return dbStructure, nil
}