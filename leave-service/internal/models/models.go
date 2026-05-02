package models

import "time"

type LeaveType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	MaxDays int    `json:"maxDays"`
}

type Employee struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ManagerID *string `json:"managerId,omitempty"`
}

type LeaveBalance struct {
	LeaveTypeID   string `json:"leaveTypeId"`
	LeaveTypeName string `json:"leaveTypeName"`
	Allocated     int    `json:"allocated"`
	Used          int    `json:"used"`
	Remaining     int    `json:"remaining"`
}

type LeaveRequest struct {
	ID              string     `json:"id"`
	EmployeeID      string     `json:"employeeId"`
	EmployeeName    string     `json:"employeeName"`
	LeaveTypeID     string     `json:"leaveTypeId"`
	LeaveTypeName   string     `json:"leaveTypeName"`
	StartDate       string     `json:"startDate"`
	EndDate         string     `json:"endDate"`
	NumberOfDays    int        `json:"numberOfDays"`
	Reason          string     `json:"reason"`
	Status          string     `json:"status"`
	ManagerID       *string    `json:"managerId,omitempty"`
	ManagerComments *string    `json:"managerComments,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type SubmitLeaveRequest struct {
	EmployeeID  string `json:"employeeId"`
	LeaveTypeID string `json:"leaveTypeId"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Reason      string `json:"reason"`
}

type ApproveRejectRequest struct {
	ManagerID string  `json:"managerId"`
	Comments  *string `json:"comments,omitempty"`
}
