package main

import (
	"fmt"
	"net/http"
)

type Vehicle struct {
	Plate      string
	Maker      string
	Model      string
	Year       string
	AssignedTo string
}

var vehicles = []Vehicle{
	{"ABC-123", "Chevrolet", "Spark", "2017", "Andrés"},
	{"DEF-456", "Chevrolet", "Chevy", "2010", "Miguel"},
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hola, CIETEL!\n")

	for _, v := range vehicles {
		fmt.Fprintf(w, "Vehicle: %s\t%s\t%s\t%s\n", v.Maker, v.Model, v.Year, v.Plate)
	}
	fmt.Println("Replacing on template...")
}

func main() {
	http.HandleFunc("/vehiculos", hello)

	http.ListenAndServe(":8080", nil)
}
