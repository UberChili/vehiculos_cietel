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

var Makers = []string{"chevrolet", "hyundai", "ford"}
var Models = []string{"spark", "chevy", "ikon", "atos", "courier", "tornado"}

type Record struct {
	ID          int
	VehicleID   int
	DateShort   string
	Type        string
	Description string
	Cost        string
}
