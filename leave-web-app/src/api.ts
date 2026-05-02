import axios from 'axios';
import { LeaveType, LeaveBalance, LeaveRequest, CreateLeaveRequest, ReviewLeaveRequest } from './types';

const BASE_URL: string =
  (window as unknown as { RUNTIME_BACKEND_API_URL?: string }).RUNTIME_BACKEND_API_URL ||
  (import.meta as unknown as { env: { VITE_API_BASE_URL?: string } }).env.VITE_API_BASE_URL ||
  '/api';

const api = axios.create({
  baseURL: BASE_URL,
  headers: { 'Content-Type': 'application/json' }
});

export async function getLeaveTypes(): Promise<LeaveType[]> {
  const res = await api.get<LeaveType[]>('/leave-types');
  return res.data;
}

export async function getLeaveBalance(employeeId: string): Promise<LeaveBalance[]> {
  const res = await api.get<LeaveBalance[]>(`/employees/${employeeId}/balance`);
  return res.data;
}

export async function createLeaveRequest(data: CreateLeaveRequest): Promise<LeaveRequest> {
  const res = await api.post<LeaveRequest>('/leave-requests', data);
  return res.data;
}

export async function getLeaveRequests(params?: {
  employeeId?: string;
  status?: string;
  startDate?: string;
  endDate?: string;
}): Promise<LeaveRequest[]> {
  const res = await api.get<LeaveRequest[]>('/leave-requests', { params });
  return res.data;
}

export async function approveLeaveRequest(requestId: string, data: ReviewLeaveRequest): Promise<LeaveRequest> {
  const res = await api.post<LeaveRequest>(`/leave-requests/${requestId}/approve`, data);
  return res.data;
}

export async function rejectLeaveRequest(requestId: string, data: ReviewLeaveRequest): Promise<LeaveRequest> {
  const res = await api.post<LeaveRequest>(`/leave-requests/${requestId}/reject`, data);
  return res.data;
}

export async function healthCheck(): Promise<{ status: string }> {
  const res = await api.get<{ status: string }>('/health');
  return res.data;
}
