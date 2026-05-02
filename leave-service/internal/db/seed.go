package db

import (
	"database/sql"
	"log"
)

type seedData struct {
	leaveTypeID string
	name        string
	maxDays     int
}

var leaveTypes = []seedData{
	{"a1b2c3d4-0001-0001-0001-000000000001", "Annual", 20},
	{"a1b2c3d4-0002-0002-0002-000000000002", "Sick", 10},
	{"a1b2c3d4-0003-0003-0003-000000000003", "Personal", 5},
}

type empData struct {
	id        string
	name      string
	managerID *string
}

var mgr1ID = "e0000001-0001-0001-0001-000000000001"
var mgr2ID = "e0000002-0002-0002-0002-000000000002"

var employees = []empData{
	{mgr1ID, "Alice Johnson", nil},
	{mgr2ID, "Bob Williams", nil},
	{"e0000003-0003-0003-0003-000000000003", "Charlie Brown", &mgr1ID},
	{"e0000004-0004-0004-0004-000000000004", "Diana Prince", &mgr1ID},
	{"e0000005-0005-0005-0005-000000000005", "Eve Davis", &mgr2ID},
}

func SeedData(db *sql.DB) error {
	// Check if already seeded
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM leave_types").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		log.Println("Database already seeded, skipping.")
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert leave types
	for _, lt := range leaveTypes {
		_, err := tx.Exec(
			`INSERT INTO leave_types (id, name, max_days) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			lt.leaveTypeID, lt.name, lt.maxDays,
		)
		if err != nil {
			return err
		}
	}

	// Insert employees (managers first)
	for _, emp := range employees {
		if emp.managerID == nil {
			_, err := tx.Exec(
				`INSERT INTO employees (id, name, manager_id) VALUES ($1, $2, NULL) ON CONFLICT DO NOTHING`,
				emp.id, emp.name,
			)
			if err != nil {
				return err
			}
		}
	}
	for _, emp := range employees {
		if emp.managerID != nil {
			_, err := tx.Exec(
				`INSERT INTO employees (id, name, manager_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				emp.id, emp.name, emp.managerID,
			)
			if err != nil {
				return err
			}
		}
	}

	// Insert leave balances for all employees
	balanceID := 1
	for _, emp := range employees {
		for _, lt := range leaveTypes {
			_, err := tx.Exec(
				`INSERT INTO leave_balances (id, employee_id, leave_type_id, allocated, used)
				 VALUES (gen_random_uuid(), $1, $2, $3, 0) ON CONFLICT DO NOTHING`,
				emp.id, lt.leaveTypeID, lt.maxDays,
			)
			if err != nil {
				// fallback if gen_random_uuid not available
				_, err = tx.Exec(
					`INSERT INTO leave_balances (id, employee_id, leave_type_id, allocated, used)
					 VALUES ($1, $2, $3, $4, 0) ON CONFLICT DO NOTHING`,
					generateBalanceID(balanceID), emp.id, lt.leaveTypeID, lt.maxDays,
				)
				if err != nil {
					return err
				}
			}
			balanceID++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Database seeded successfully.")
	return nil
}

func generateBalanceID(n int) string {
	return "b0000000-0000-0000-0000-" + padLeft(n, 12)
}

func padLeft(n, width int) string {
	s := ""
	num := n
	for i := 0; i < width; i++ {
		s = string(rune('0'+num%10)) + s
		num /= 10
	}
	return s
}
