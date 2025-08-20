package model

type Mode string

const (
	Prod Mode = "prod"
	Dev  Mode = "dev"
)

type Status string

const (
	Dead         Status = "Dead"
	Closed       Status = "Closed"
	Refranchised Status = "Refranchised"
	Open         Status = "Open"
	New          Status = "New"
	PreOpening   Status = "PreOpening"
	Undefined    Status = "Undefined"
)

type Store struct {
	Number          int
	Name            string
	Address         string
	Mall            string
	Franchise       string
	Brand           string
	Format          string
	Status          Status
	TemporaryClosed bool
}
