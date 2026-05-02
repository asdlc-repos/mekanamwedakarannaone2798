import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { getLeaveRequests, approveLeaveRequest, rejectLeaveRequest } from '../api'
import { LeaveRequest } from '../types'

type FilterStatus = 'pending' | 'all'

function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`status-badge status-${status}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  )
}

interface ReviewModalProps {
  request: LeaveRequest;
  action: 'approve' | 'reject';
  onConfirm: (comment: string) => void;
  onCancel: () => void;
  loading: boolean;
}

function ReviewModal({ request, action, onConfirm, onCancel, loading }: ReviewModalProps) {
  const [comment, setComment] = useState('')

  return (
    <div className="modal-overlay">
      <div className="modal">
        <h3 className="modal-title">
          {action === 'approve' ? 'Approve' : 'Reject'} Leave Request
        </h3>
        <p style={{ marginBottom: 16, color: '#555', fontSize: 14 }}>
          <strong>{request.employeeName || request.employeeId}</strong> — {request.leaveTypeName || request.leaveTypeId}<br />
          <span style={{ color: '#888' }}>{request.startDate} to {request.endDate}</span>
        </p>
        <div className="form-group">
          <label className="form-label">Comment (optional)</label>
          <textarea
            className="form-control"
            rows={3}
            value={comment}
            placeholder="Add a comment for the employee..."
            onChange={e => setComment(e.target.value)}
          />
        </div>
        <div className="modal-actions">
          <button className="btn btn-secondary" onClick={onCancel} disabled={loading}>
            Cancel
          </button>
          <button
            className={`btn ${action === 'approve' ? 'btn-success' : 'btn-danger'}`}
            onClick={() => onConfirm(comment)}
            disabled={loading}
          >
            {loading ? 'Processing...' : (action === 'approve' ? 'Approve' : 'Reject')}
          </button>
        </div>
      </div>
    </div>
  )
}

export default function ManagerPage() {
  const [requests, setRequests] = useState<LeaveRequest[]>([])
  const [loading, setLoading] = useState(true)
  const [filterStatus, setFilterStatus] = useState<FilterStatus>('pending')
  const [filterEmployee, setFilterEmployee] = useState('')
  const [filterStart, setFilterStart] = useState('')
  const [filterEnd, setFilterEnd] = useState('')
  const [reviewModal, setReviewModal] = useState<{
    request: LeaveRequest;
    action: 'approve' | 'reject';
  } | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  const fetchRequests = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, string> = {}
      if (filterStatus === 'pending') params.status = 'pending'
      if (filterEmployee.trim()) params.employeeId = filterEmployee.trim()
      if (filterStart) params.startDate = filterStart
      if (filterEnd) params.endDate = filterEnd
      const data = await getLeaveRequests(params)
      setRequests(data)
    } catch {
      toast.error('Failed to load leave requests')
    } finally {
      setLoading(false)
    }
  }, [filterStatus, filterEmployee, filterStart, filterEnd])

  useEffect(() => {
    fetchRequests()
  }, [fetchRequests])

  async function handleReview(comment: string) {
    if (!reviewModal) return
    setActionLoading(true)
    try {
      if (reviewModal.action === 'approve') {
        await approveLeaveRequest(reviewModal.request.id, { comment: comment || undefined })
        toast.success('Leave request approved')
      } else {
        await rejectLeaveRequest(reviewModal.request.id, { comment: comment || undefined })
        toast.success('Leave request rejected')
      }
      setReviewModal(null)
      fetchRequests()
    } catch {
      toast.error(`Failed to ${reviewModal.action} request`)
    } finally {
      setActionLoading(false)
    }
  }

  return (
    <div>
      <div className="card">
        <div className="section-header">
          <h3>Leave Requests</h3>
          <button className="btn btn-secondary btn-sm" onClick={fetchRequests} disabled={loading}>
            Refresh
          </button>
        </div>

        <div className="filters">
          <div className="form-group">
            <label className="form-label">Status</label>
            <select
              className="form-control"
              value={filterStatus}
              onChange={e => setFilterStatus(e.target.value as FilterStatus)}
            >
              <option value="pending">Pending Only</option>
              <option value="all">All Requests</option>
            </select>
          </div>
          <div className="form-group">
            <label className="form-label">Employee ID</label>
            <input
              type="text"
              className="form-control"
              placeholder="Filter by employee ID..."
              value={filterEmployee}
              onChange={e => setFilterEmployee(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">From Date</label>
            <input
              type="date"
              className="form-control"
              value={filterStart}
              onChange={e => setFilterStart(e.target.value)}
            />
          </div>
          <div className="form-group">
            <label className="form-label">To Date</label>
            <input
              type="date"
              className="form-control"
              value={filterEnd}
              onChange={e => setFilterEnd(e.target.value)}
            />
          </div>
          <button
            className="btn btn-secondary btn-sm"
            style={{ alignSelf: 'flex-end', marginBottom: 0 }}
            onClick={() => { setFilterEmployee(''); setFilterStart(''); setFilterEnd('') }}
          >
            Clear
          </button>
        </div>

        {loading ? (
          <div className="loading">
            <div className="spinner" />
            Loading requests...
          </div>
        ) : requests.length === 0 ? (
          <div className="empty-state">
            <p>No leave requests found.</p>
          </div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Employee</th>
                  <th>Leave Type</th>
                  <th>Start Date</th>
                  <th>End Date</th>
                  <th>Reason</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {requests.map(r => (
                  <tr key={r.id}>
                    <td>
                      <div style={{ fontWeight: 500 }}>{r.employeeName || r.employeeId}</div>
                      {r.employeeName && <div style={{ fontSize: 12, color: '#888' }}>{r.employeeId}</div>}
                    </td>
                    <td>{r.leaveTypeName || r.leaveTypeId}</td>
                    <td>{r.startDate}</td>
                    <td>{r.endDate}</td>
                    <td style={{ maxWidth: 200, wordBreak: 'break-word' }}>{r.reason}</td>
                    <td><StatusBadge status={r.status} /></td>
                    <td>
                      {r.status === 'pending' ? (
                        <div className="action-group">
                          <button
                            className="btn btn-success btn-sm"
                            onClick={() => setReviewModal({ request: r, action: 'approve' })}
                          >
                            Approve
                          </button>
                          <button
                            className="btn btn-danger btn-sm"
                            onClick={() => setReviewModal({ request: r, action: 'reject' })}
                          >
                            Reject
                          </button>
                        </div>
                      ) : (
                        <span style={{ color: '#999', fontSize: 13 }}>
                          {r.managerComment || '—'}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {reviewModal && (
        <ReviewModal
          request={reviewModal.request}
          action={reviewModal.action}
          onConfirm={handleReview}
          onCancel={() => setReviewModal(null)}
          loading={actionLoading}
        />
      )}
    </div>
  )
}
