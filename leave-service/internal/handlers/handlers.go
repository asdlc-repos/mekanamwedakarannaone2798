package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/asdlc-repos/mekanamwedakarannaone2/leave-service/internal/models"
	"github.com/google/uuid"
)

type Handler struct {
	DB *sql.DB
}

func New(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) http.Handler {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/api/leave-types", h.handleLeaveTypes)
	mux.HandleFunc("/api/employees/", h.handleEmployeeRoutes)
	mux.HandleFunc("/api/leave-requests", h.handleLeaveRequests)
	mux.HandleFunc("/api/leave-requests/", h.handleLeaveRequestByID)
	return corsMiddleware(mux)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

func (h *Handler) handleLeaveTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	rows, err := h.DB.Query(`SELECT id, name, max_days FROM leave_types ORDER BY name`)
	if err != nil {
		log.Printf("error querying leave types: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	types := []models.LeaveType{}
	for rows.Next() {
		var lt models.LeaveType
		if err := rows.Scan(&lt.ID, &lt.Name, &lt.MaxDays); err != nil {
			log.Printf("error scanning leave type: %v", err)
			continue
		}
		types = append(types, lt)
	}
	writeJSON(w, http.StatusOK, types)
}

// Route: /api/employees/{employeeId}/balance
func (h *Handler) handleEmployeeRoutes(w http.ResponseWriter, r *http.Request) {
	// path: /api/employees/{employeeId}/balance
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/employees/"), "/")
	if len(parts) == 2 && parts[1] == "balance" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleGetBalance(w, r, parts[0])
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}

func (h *Handler) handleGetBalance(w http.ResponseWriter, r *http.Request, employeeID string) {
	// Check employee exists
	var empName string
	err := h.DB.QueryRow(`SELECT name FROM employees WHERE id = $1`, employeeID).Scan(&empName)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "employee not found")
		return
	} else if err != nil {
		log.Printf("error querying employee: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	rows, err := h.DB.Query(`
		SELECT lb.leave_type_id, lt.name, lb.allocated, lb.used
		FROM leave_balances lb
		JOIN leave_types lt ON lt.id = lb.leave_type_id
		WHERE lb.employee_id = $1
		ORDER BY lt.name
	`, employeeID)
	if err != nil {
		log.Printf("error querying leave balances: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	balances := []models.LeaveBalance{}
	for rows.Next() {
		var b models.LeaveBalance
		if err := rows.Scan(&b.LeaveTypeID, &b.LeaveTypeName, &b.Allocated, &b.Used); err != nil {
			log.Printf("error scanning balance: %v", err)
			continue
		}
		b.Remaining = b.Allocated - b.Used
		balances = append(balances, b)
	}
	writeJSON(w, http.StatusOK, balances)
}

func (h *Handler) handleLeaveRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetLeaveRequests(w, r)
	case http.MethodPost:
		h.handleSubmitLeaveRequest(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleGetLeaveRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	employeeID := q.Get("employeeId")
	managerID := q.Get("managerId")
	status := q.Get("status")

	query := `
		SELECT lr.id, lr.employee_id, e.name, lr.leave_type_id, lt.name,
		       TO_CHAR(lr.start_date, 'YYYY-MM-DD'), TO_CHAR(lr.end_date, 'YYYY-MM-DD'),
		       lr.number_of_days, COALESCE(lr.reason, ''), lr.status,
		       lr.manager_id, lr.manager_comments, lr.created_at, lr.updated_at
		FROM leave_requests lr
		JOIN employees e ON e.id = lr.employee_id
		JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if employeeID != "" {
		query += " AND lr.employee_id = $" + itoa(argIdx)
		args = append(args, employeeID)
		argIdx++
	}
	if managerID != "" {
		// Filter by employees whose manager is this manager
		query += " AND e.manager_id = $" + itoa(argIdx)
		args = append(args, managerID)
		argIdx++
	}
	if status != "" {
		query += " AND lr.status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY lr.created_at DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		log.Printf("error querying leave requests: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	requests := []models.LeaveRequest{}
	for rows.Next() {
		lr, err := scanLeaveRequest(rows)
		if err != nil {
			log.Printf("error scanning leave request: %v", err)
			continue
		}
		requests = append(requests, lr)
	}
	writeJSON(w, http.StatusOK, requests)
}

func (h *Handler) handleSubmitLeaveRequest(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitLeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EmployeeID == "" || req.LeaveTypeID == "" || req.StartDate == "" || req.EndDate == "" {
		writeError(w, http.StatusBadRequest, "employeeId, leaveTypeId, startDate, and endDate are required")
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startDate format, use YYYY-MM-DD")
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid endDate format, use YYYY-MM-DD")
		return
	}
	if endDate.Before(startDate) {
		writeError(w, http.StatusBadRequest, "endDate must be >= startDate")
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	if startDate.Before(today) {
		writeError(w, http.StatusBadRequest, "startDate cannot be in the past")
		return
	}

	numberOfDays := businessDays(startDate, endDate)
	if numberOfDays <= 0 {
		writeError(w, http.StatusBadRequest, "leave request must include at least one business day")
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("error starting transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer tx.Rollback()

	// Check employee exists
	var empName string
	err = tx.QueryRow(`SELECT name FROM employees WHERE id = $1`, req.EmployeeID).Scan(&empName)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusBadRequest, "employee not found")
		return
	} else if err != nil {
		log.Printf("error querying employee: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Check leave type exists
	var ltName string
	err = tx.QueryRow(`SELECT name FROM leave_types WHERE id = $1`, req.LeaveTypeID).Scan(&ltName)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusBadRequest, "leave type not found")
		return
	} else if err != nil {
		log.Printf("error querying leave type: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Check and update balance atomically
	var allocated, used int
	err = tx.QueryRow(`
		SELECT allocated, used FROM leave_balances
		WHERE employee_id = $1 AND leave_type_id = $2
		FOR UPDATE
	`, req.EmployeeID, req.LeaveTypeID).Scan(&allocated, &used)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusBadRequest, "no leave balance found for this employee and leave type")
		return
	} else if err != nil {
		log.Printf("error querying balance: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	remaining := allocated - used
	if remaining < numberOfDays {
		writeError(w, http.StatusBadRequest, "insufficient leave balance")
		return
	}

	// Deduct balance
	_, err = tx.Exec(`
		UPDATE leave_balances SET used = used + $1
		WHERE employee_id = $2 AND leave_type_id = $3
	`, numberOfDays, req.EmployeeID, req.LeaveTypeID)
	if err != nil {
		log.Printf("error updating balance: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Create request
	id := uuid.New().String()
	now := time.Now()
	_, err = tx.Exec(`
		INSERT INTO leave_requests (id, employee_id, leave_type_id, start_date, end_date, number_of_days, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, $9)
	`, id, req.EmployeeID, req.LeaveTypeID, startDate, endDate, numberOfDays, req.Reason, now, now)
	if err != nil {
		log.Printf("error inserting leave request: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("error committing transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	lr := models.LeaveRequest{
		ID:            id,
		EmployeeID:    req.EmployeeID,
		EmployeeName:  empName,
		LeaveTypeID:   req.LeaveTypeID,
		LeaveTypeName: ltName,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		NumberOfDays:  numberOfDays,
		Reason:        req.Reason,
		Status:        "pending",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	writeJSON(w, http.StatusCreated, lr)
}

// Route: /api/leave-requests/{requestId} and sub-paths
func (h *Handler) handleLeaveRequestByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/leave-requests/")
	parts := strings.SplitN(path, "/", 2)
	requestID := parts[0]

	if len(parts) == 1 {
		// /api/leave-requests/{requestId}
		if r.Method == http.MethodGet {
			h.handleGetLeaveRequestByID(w, r, requestID)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	action := parts[1]
	switch action {
	case "approve":
		if r.Method == http.MethodPost {
			h.handleApproveReject(w, r, requestID, "approved")
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	case "reject":
		if r.Method == http.MethodPost {
			h.handleApproveReject(w, r, requestID, "rejected")
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) handleGetLeaveRequestByID(w http.ResponseWriter, r *http.Request, requestID string) {
	row := h.DB.QueryRow(`
		SELECT lr.id, lr.employee_id, e.name, lr.leave_type_id, lt.name,
		       TO_CHAR(lr.start_date, 'YYYY-MM-DD'), TO_CHAR(lr.end_date, 'YYYY-MM-DD'),
		       lr.number_of_days, COALESCE(lr.reason, ''), lr.status,
		       lr.manager_id, lr.manager_comments, lr.created_at, lr.updated_at
		FROM leave_requests lr
		JOIN employees e ON e.id = lr.employee_id
		JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE lr.id = $1
	`, requestID)

	lr, err := scanLeaveRequestRow(row)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "leave request not found")
		return
	} else if err != nil {
		log.Printf("error querying leave request: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, lr)
}

func (h *Handler) handleApproveReject(w http.ResponseWriter, r *http.Request, requestID, action string) {
	var req models.ApproveRejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ManagerID == "" {
		writeError(w, http.StatusBadRequest, "managerId is required")
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("error starting transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer tx.Rollback()

	// Get the leave request with lock
	var lr struct {
		employeeID  string
		leaveTypeID string
		numberOfDays int
		status      string
	}
	err = tx.QueryRow(`
		SELECT employee_id, leave_type_id, number_of_days, status
		FROM leave_requests WHERE id = $1 FOR UPDATE
	`, requestID).Scan(&lr.employeeID, &lr.leaveTypeID, &lr.numberOfDays, &lr.status)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "leave request not found")
		return
	} else if err != nil {
		log.Printf("error querying leave request: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if lr.status != "pending" {
		writeError(w, http.StatusBadRequest, "leave request has already been processed")
		return
	}

	// Verify manager authorization
	var managerIDFromDB sql.NullString
	err = tx.QueryRow(`SELECT manager_id FROM employees WHERE id = $1`, lr.employeeID).Scan(&managerIDFromDB)
	if err != nil {
		log.Printf("error querying employee manager: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !managerIDFromDB.Valid || managerIDFromDB.String != req.ManagerID {
		writeError(w, http.StatusForbidden, "manager not authorized to process this request")
		return
	}

	// If rejecting, restore balance
	if action == "rejected" {
		_, err = tx.Exec(`
			UPDATE leave_balances SET used = used - $1
			WHERE employee_id = $2 AND leave_type_id = $3
		`, lr.numberOfDays, lr.employeeID, lr.leaveTypeID)
		if err != nil {
			log.Printf("error restoring balance: %v", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	now := time.Now()
	_, err = tx.Exec(`
		UPDATE leave_requests
		SET status = $1, manager_id = $2, manager_comments = $3, updated_at = $4
		WHERE id = $5
	`, action, req.ManagerID, req.Comments, now, requestID)
	if err != nil {
		log.Printf("error updating leave request: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("error committing transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Return updated request
	h.handleGetLeaveRequestByID(w, r, requestID)
}

// Helper: scan a leave request from rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanLeaveRequestRow(row rowScanner) (models.LeaveRequest, error) {
	var lr models.LeaveRequest
	var managerID sql.NullString
	var managerComments sql.NullString
	err := row.Scan(
		&lr.ID, &lr.EmployeeID, &lr.EmployeeName, &lr.LeaveTypeID, &lr.LeaveTypeName,
		&lr.StartDate, &lr.EndDate, &lr.NumberOfDays, &lr.Reason, &lr.Status,
		&managerID, &managerComments, &lr.CreatedAt, &lr.UpdatedAt,
	)
	if err != nil {
		return lr, err
	}
	if managerID.Valid {
		lr.ManagerID = &managerID.String
	}
	if managerComments.Valid {
		lr.ManagerComments = &managerComments.String
	}
	return lr, nil
}

func scanLeaveRequest(rows *sql.Rows) (models.LeaveRequest, error) {
	var lr models.LeaveRequest
	var managerID sql.NullString
	var managerComments sql.NullString
	err := rows.Scan(
		&lr.ID, &lr.EmployeeID, &lr.EmployeeName, &lr.LeaveTypeID, &lr.LeaveTypeName,
		&lr.StartDate, &lr.EndDate, &lr.NumberOfDays, &lr.Reason, &lr.Status,
		&managerID, &managerComments, &lr.CreatedAt, &lr.UpdatedAt,
	)
	if err != nil {
		return lr, err
	}
	if managerID.Valid {
		lr.ManagerID = &managerID.String
	}
	if managerComments.Valid {
		lr.ManagerComments = &managerComments.String
	}
	return lr, nil
}

// businessDays counts business days (Mon-Fri) between two dates (inclusive)
func businessDays(start, end time.Time) int {
	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
	}
	return count
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 3)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
