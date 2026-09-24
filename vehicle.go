package main

type Vehicle struct {
	ID             int
	Plate          string
	Maker          string
	Model          string
	Year           string
	LastService    string
	LastRepair     string
	AssignedTo     *int
	AssignedToName string
	Location       string
	Records        []Record
	PhotoURL       string
}

// func (v Vehicle)

// Every maker in the fleet and its models, lowercase like they're stored.
var ModelsByMaker = map[string][]string{
	"chevrolet": {"spark", "chevy", "tornado"},
	"hyundai":   {"atos"},
	"ford":      {"courier", "ikon"},
}

// Every city we work in, grouped by state abbreviation. A vehicle's location
// is stored as "City, State", e.g. "Morelia, Mich.".
var CitiesByState = map[string][]string{
	"Mich.": {"Morelia", "Tangancícuaro", "Pátzcuaro", "La Piedad", "Sahuayo"},
	"Jal.":  {"Puerto Vallarta"},
	"Gto.":  {"San José Iturbide"},
}

type Record struct {
	ID          int
	VehicleID   int
	DateShort   string
	Type        string
	Description string
	Cost        *int64 // in cents, so $1850.50 is 185050. nil when no cost was entered
}
