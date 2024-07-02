package entities

type UserInterface interface {
	GetID() int
	GetEmail() string
	GetPassword() string
}

type ChirpInterface interface {
	GetID() int
	GetBody() string
}
