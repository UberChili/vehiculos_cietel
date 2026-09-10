package main

type Vehicle struct {
	ID          int
	Maker       string
	Model       string
	Year        string
	Plate       string
	LastService string
	LastRepair  string
	AssignedTo  string
	Location    string
	Records     []Record
	PhotoURL    string
}

var Makers = []string{"chevrolet", "hyundai", "ford"}
var Models = []string{"spark", "chevy", "ikon", "atos"}

type Record struct {
	ID          int
	DateShort   string
	Type        string
	Description string
	Cost        string
}
