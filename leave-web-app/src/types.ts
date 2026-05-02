export type UserRole = 'employee' | 'manager';

export interface User {
  id: string;
  name: string;
  role: UserRole;
}

export interface LeaveType {
  id: string;
  name: string;
  description?: string;
}

export interface LeaveBalance {
  leaveTypeId: string;
  leaveTypeName: string;
  total: number;
  used: number;
  remaining: number;
}

export interface LeaveRequest {
  id: string;
  employeeId: string;
  employeeName?: string;
  leaveTypeId: string;
  leaveTypeName?: string;
  startDate: string;
  endDate: string;
  reason: string;
  status: 'pending' | 'approved' | 'rejected';
  managerComment?: string;
  createdAt?: string;
}

export interface CreateLeaveRequest {
  employeeId: string;
  leaveTypeId: string;
  startDate: string;
  endDate: string;
  reason: string;
}

export interface ReviewLeaveRequest {
  comment?: string;
}
