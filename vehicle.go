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

type Record struct {
	ID          int
	VehicleID   int
	DateShort   string
	Type        string
	Description string
	Cost        string
}
