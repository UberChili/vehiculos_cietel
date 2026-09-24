package main

type Vehicle struct {
	ID             int
	Plate          string
	Maker          string
	Model          string
	Year           string
	LastService    string
	LastRepair     string
	LastBandas     string // last record mentioning "banda(s)" in its description
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

	// Data for the future reports (see the reports notes). All optional except Cause on repairs.
	OdometerKm     *int   // nil when not read or when the odometer is broken, never 0 for "unknown"
	OdometerBroken bool   // "El odómetro no funciona" was checked
	DowntimeDays   *int   // days the vehicle couldn't be used. nil when not filled in
	Cause          string // only for repairs, one of RepairCauses. "" for services
	TechnicianID   *int   // who had the vehicle when the record was created (set automatically)
	TechnicianName string // for display. "" when nobody had it or the technician was deleted
}

// Records dated more than this many days ago don't get the technician
// snapshot (Record.TechnicianID), because they're old history being entered
// late and the vehicle may have had someone else back then.
const TechnicianSnapshotMaxDays = 30

// Why a repair happened. Needed to tell apart an old car wearing out from a
// technician who doesn't take care of it.
var RepairCauses = []string{"Desgaste normal", "Mal uso o descuido", "Accidente o golpe", "No se sabe"}

// OdometerReading is one km reading of a vehicle, used by the record forms to
// warn when a new reading doesn't fit the others.
type OdometerReading struct {
	Date string // ISO, as stored
	Km   int
}
