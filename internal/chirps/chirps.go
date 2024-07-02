package chirps

import (
	"errors"
	"fmt"

	"github.com/Groskilled/Chirpy/internal/database"
	"github.com/Groskilled/Chirpy/internal/entities"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func (c Chirp) GetID() int {
	return c.Id
}

func (c Chirp) GetBody() string {
	return c.Body
}

func CreateChirp(db database.DB, id int, body string) (Chirp, error) {
	newChirp := Chirp{
		Id:   id,
		Body: body,
	}
	dbStruct, err := db.LoadDB()
	if err != nil {
		fmt.Printf("Could not load DB. Error: %s\n", err)
	}
	if dbStruct.Chirps == nil {
		dbStruct.Chirps = make(map[int]entities.ChirpInterface)
	}
	dbStruct.Chirps[newChirp.Id] = newChirp
	err = db.WriteDB(dbStruct)
	if err != nil {
		fmt.Printf("Error writing to DB: %s\n", err)
	}
	return newChirp, nil
}

func GetChirps(db database.DB) ([]Chirp, error) {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return nil, err
	}
	if dbStruct.Chirps == nil {
		return nil, errors.New("no chirps available")
	}
	chirps := make([]Chirp, 0, len(dbStruct.Chirps))
	for _, chirp := range dbStruct.Chirps {
		chirps = append(chirps, Chirp{chirp.GetID(), chirp.GetBody()})
	}
	return chirps, nil
}

func GetChirpById(db database.DB, id int) (Chirp, error) {
	dbStruct, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}
	if dbStruct.Chirps == nil {
		return Chirp{}, errors.New("no chirps available")
	}
	for _, chirp := range dbStruct.Chirps {
		if chirp.GetID() == id {
			return Chirp{Id: chirp.GetID(), Body: chirp.GetBody()}, nil
		}
	}
	return Chirp{}, fmt.Errorf("chirp with id %d not found", id)
}
