package store

import (
	"context"
	"fmt"
)

// demoCustomer is one seeded customer and the vehicles they own.
type demoCustomer struct {
	name     string
	phone    string
	city     string
	vehicles []demoVehicle
}

type demoVehicle struct {
	plate string
	make  string
	model string
	year  int32
	jobs  []demoJob
}

// demoJob is one repair on that vehicle. Statuses are mixed on purpose, so the
// job list shows every badge rather than four rows of the same one.
type demoJob struct {
	description string
	status      string
	labour      int32
}

type demoPart struct {
	reference string
	name      string
	millimes  int32
	quantity  int32
}

// Seed replaces everything with demo data.
//
// It is destructive on purpose: a seeder that appends leaves you with four copies
// of the same customer after four runs. main refuses to call it outside
// development, which is the guard that matters.
func (s *Store) Seed(ctx context.Context) error {
	// One statement: CASCADE handles the foreign key from vehicles, and RESTART
	// IDENTITY resets the id counters so seeded ids are the same every time.
	if _, err := s.pool.Exec(ctx, `TRUNCATE customers, vehicles, parts, service_jobs RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("clear existing data: %w", err)
	}

	for _, c := range demoCustomers {
		customer, err := s.CreateCustomer(ctx, c.name, c.phone, c.city)
		if err != nil {
			return fmt.Errorf("seed customer %s: %w", c.name, err)
		}

		for _, v := range c.vehicles {
			vehicle, err := s.CreateVehicle(ctx, customer.ID, v.plate, v.make, v.model, v.year)
			if err != nil {
				return fmt.Errorf("seed vehicle %s: %w", v.plate, err)
			}

			for _, j := range v.jobs {
				if _, err := s.CreateJob(ctx, vehicle.ID, j.description, j.status, j.labour); err != nil {
					return fmt.Errorf("seed job for %s: %w", v.plate, err)
				}
			}
		}
	}

	for _, p := range demoParts {
		if _, err := s.CreatePart(ctx, p.reference, p.name, p.millimes, p.quantity); err != nil {
			return fmt.Errorf("seed part %s: %w", p.reference, err)
		}
	}

	return nil
}

var demoCustomers = []demoCustomer{
	{
		name: "Mohamed Ben Salah", phone: "+216 20 145 872", city: "Tunis",
		vehicles: []demoVehicle{
			{"123 TUN 4567", "Renault", "Clio", 2019, []demoJob{
				{"Grinding noise from the front when braking", JobInProgress, 45000},
				{"Annual service and oil change", JobCompleted, 60000},
			}},
			{"98 TUN 1204", "Peugeot", "Partner", 2016, []demoJob{
				{"Clutch slipping on hills", JobReceived, 0},
			}},
		},
	},
	{
		name: "Amira Trabelsi", phone: "+216 55 903 214", city: "Sfax",
		vehicles: []demoVehicle{{"204 TUN 8891", "Volkswagen", "Golf", 2021, []demoJob{
			{"Air conditioning blows warm", JobInProgress, 35000},
		}}},
	},
	{
		name: "Youssef Gharbi", phone: "+216 98 476 130", city: "Sousse",
		vehicles: []demoVehicle{
			{"77 TUN 3345", "Dacia", "Logan", 2018, []demoJob{
				{"Replace front brake pads and discs", JobCompleted, 80000},
				{"Customer changed their mind about the body work", JobCancelled, 0},
			}},
			{"310 TUN 5520", "Citroën", "C3", 2022, nil},
		},
	},
	{
		name: "Salma Bouazizi", phone: "+216 22 318 605", city: "Ariana",
		vehicles: []demoVehicle{{"441 TUN 7712", "Hyundai", "i10", 2020, []demoJob{
			{"Warning light on the dashboard, needs diagnostics", JobReceived, 0},
		}}},
	},
	{
		name: "Karim Jebali", phone: "+216 71 260 449", city: "Bizerte",
		vehicles: []demoVehicle{{"158 TUN 9034", "Fiat", "Panda", 2015, []demoJob{
			{"Timing belt replacement at 120,000 km", JobInProgress, 120000},
		}}},
	},
	{
		name: "Nadia Mansouri", phone: "+216 24 771 508", city: "Nabeul",
		// No job: a car with no service history yet is worth seeing too.
		vehicles: []demoVehicle{{"502 TUN 1187", "Toyota", "Yaris", 2023, nil}},
	},
	{
		name: "Hichem Aouadi", phone: "+216 53 640 219", city: "Monastir",
		// No vehicle: the empty state on a customer page is worth seeing too.
	},
	{
		name: "Leila Chaabane", phone: "+216 29 405 736", city: "Gabès",
		vehicles: []demoVehicle{{"66 TUN 2298", "Seat", "Ibiza", 2017, []demoJob{
			{"Battery replaced, electrical check", JobCompleted, 25000},
		}}},
	},
}

// Two parts are out of stock on purpose, so the dashboard has something to show
// under "needs restocking".
var demoParts = []demoPart{
	{"BRK-001", "Brake pad set, front", 42500, 12},
	{"BRK-002", "Brake disc, front pair", 128000, 4},
	{"OIL-005", "Oil filter", 8290, 0},
	{"OIL-010", "Engine oil 5W-30, 5L", 76500, 9},
	{"AIR-003", "Air filter", 14750, 15},
	{"SPK-004", "Spark plug set", 38000, 7},
	{"WIP-006", "Wiper blade pair", 22400, 11},
	{"BAT-007", "Battery 60Ah", 245000, 3},
	{"TYR-200", "Tyre 195/65 R15", 180000, 8},
	{"TIM-011", "Timing belt kit", 310500, 0},
	{"CLU-012", "Clutch kit", 425000, 2},
	{"LMP-013", "Headlight bulb H7", 9600, 24},
}
