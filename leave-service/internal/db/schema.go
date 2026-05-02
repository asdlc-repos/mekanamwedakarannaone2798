package db

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS leave_types (
	id UUID PRIMARY KEY,
	name VARCHAR(100) NOT NULL UNIQUE,
	max_days INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS employees (
	id UUID PRIMARY KEY,
	name VARCHAR(200) NOT NULL,
	manager_id UUID REFERENCES employees(id)
);

CREATE TABLE IF NOT EXISTS leave_balances (
	id UUID PRIMARY KEY,
	employee_id UUID NOT NULL REFERENCES employees(id),
	leave_type_id UUID NOT NULL REFERENCES leave_types(id),
	allocated INTEGER NOT NULL DEFAULT 0,
	used INTEGER NOT NULL DEFAULT 0,
	UNIQUE(employee_id, leave_type_id)
);

CREATE TABLE IF NOT EXISTS leave_requests (
	id UUID PRIMARY KEY,
	employee_id UUID NOT NULL REFERENCES employees(id),
	leave_type_id UUID NOT NULL REFERENCES leave_types(id),
	start_date DATE NOT NULL,
	end_date DATE NOT NULL,
	number_of_days INTEGER NOT NULL,
	reason TEXT,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	manager_id UUID REFERENCES employees(id),
	manager_comments TEXT,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

func MigrateSchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}
	return nil
}
