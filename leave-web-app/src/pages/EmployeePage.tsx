import { useState, useEffect, useCallback } from 'react'
import toast from 'react-hot-toast'
import { useUser } from '../context/UserContext'
import { getLeaveTypes, getLeaveBalance, createLeaveRequest, getLeaveRequests } from '../api'
import { LeaveType, LeaveBalance, LeaveRequest } from '../types'

const TODAY = new Date().toISOString().split('T')[0]

type Tab = 'dashboard' | 'submit' | 'history'

function StatusBadge({ status }: { status: string }) {
  return (
    <span className={`status-badge status-${status}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  )
}

function BalanceDashboard({ balances, loading }: { balances: LeaveBalance[]; loading: boolean }) {
  const colorClasses = ['annual', 'sick', 'casual']
  if (loading) {
    return (
      <div className="loading">
        <div className="spinner" />
        Loading balances...
      </div>
    )
  }
  if (balances.length === 0) {
    return <div className="empty-state"><p>No balance data available.</p></div>
  }
  return (
    <div className="grid-3">
      {balances.map((b, i) => (
        <div key={b.leaveTypeId} className={`balance-card ${colorClasses[i % colorClasses.length]}`}>
          <div className="balance-type">{b.leaveTypeName}</div>
          <div className="balance-remaining">{b.remaining}</div>
          <div className="balance-total">of {b.total} days remaining</div>
          <div style={{ marginTop: 6, fontSize: 12, opacity: 0.8 }}>{b.used} used</div>
        </div>
      ))}
    </div>
  )
}

function SubmitForm({ leaveTypes, onSuccess }: { leaveTypes: LeaveType[]; onSuccess: () => void }) {
  const { user } = useUser()
  const [form, setForm] = useState({
    leaveTypeId: '',
    startDate: '',
    endDate: '',
    reason: ''
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)

  function validate() {
    const errs: Record<string, string> = {}
    if (!form.leaveTypeId) errs.leaveTypeId = 'Please select a leave type'
    if (!form.startDate) errs.startDate = 'Start date is required'
    if (!form.endDate) errs.endDate = 'End date is required'
    if (form.startDate && form.startDate < TODAY) errs.startDate = 'Start date cannot be in the past'
    if (form.endDate && form.startDate && form.endDate < form.startDate) errs.endDate = 'End date must be after start date'
    if (!form.reason.trim()) errs.reason = 'Reason is required'
    if (form.reason.trim().length < 5) errs.reason = 'Reason must be at least 5 characters'
    return errs
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const errs = validate()
    if (Object.keys(errs).length > 0) {
      setErrors(errs)
      return
    }
    setErrors({})
    setSubmitting(true)
    try {
      await createLeaveRequest({
        employeeId: user!.id,
        leaveTypeId: form.leaveTypeId,
        startDate: form.startDate,
        endDate: form.endDate,
        reason: form.reason.trim()
      })
      toast.success('Leave request submitted successfully!')
      setForm({ leaveTypeId: '', startDate: '', endDate: '', reason: '' })
      onSuccess()
    } catch {
      toast.error('Failed to submit leave request. Please try again.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <div className="grid-2">
        <div className="form-group">
          <label className="form-label">Leave Type *</label>
          <select
            className={`form-control ${errors.leaveTypeId ? 'error' : ''}`}
            value={form.leaveTypeId}
            onChange={e => setForm(f => ({ ...f, leaveTypeId: e.target.value }))}
          >
            <option value="">Select leave type...</option>
            {leaveTypes.map(lt => (
              <option key={lt.id} value={lt.id}>{lt.name}</option>
            ))}
          </select>
          {errors.leaveTypeId && <div className="error-text">{errors.leaveTypeId}</div>}
        </div>
        <div />

        <div className="form-group">
          <label className="form-label">Start Date *</label>
          <input
            type="date"
            className={`form-control ${errors.startDate ? 'error' : ''}`}
            value={form.startDate}
            min={TODAY}
            onChange={e => setForm(f => ({ ...f, startDate: e.target.value }))}
          />
          {errors.startDate && <div className="error-text">{errors.startDate}</div>}
        </div>

        <div className="form-group">
          <label className="form-label">End Date *</label>
          <input
            type="date"
            className={`form-control ${errors.endDate ? 'error' : ''}`}
            value={form.endDate}
            min={form.startDate || TODAY}
            onChange={e => setForm(f => ({ ...f, endDate: e.target.value }))}
          />
          {errors.endDate && <div className="error-text">{errors.endDate}</div>}
        </div>
      </div>

      <div className="form-group">
        <label className="form-label">Reason *</label>
        <textarea
          className={`form-control ${errors.reason ? 'error' : ''}`}
          rows={4}
          value={form.reason}
          placeholder="Please describe the reason for your leave request..."
          onChange={e => setForm(f => ({ ...f, reason: e.target.value }))}
        />
        {errors.reason && <div className="error-text">{errors.reason}</div>}
      </div>

      <button type="submit" className="btn btn-primary" disabled={submitting}>
        {submitting ? 'Submitting...' : 'Submit Request'}
      </button>
    </form>
  )
}

function LeaveHistory({ requests, loading }: { requests: LeaveRequest[]; loading: boolean }) {
  if (loading) {
    return (
      <div className="loading">
        <div className="spinner" />
        Loading requests...
      </div>
    )
  }
  if (requests.length === 0) {
    return <div className="empty-state"><p>No leave requests found.</p></div>
  }
  return (
    <div className="table-container">
      <table>
        <thead>
          <tr>
            <th>Leave Type</th>
            <th>Start Date</th>
            <th>End Date</th>
            <th>Reason</th>
            <th>Status</th>
            <th>Manager Comment</th>
          </tr>
        </thead>
        <tbody>
          {requests.map(r => (
            <tr key={r.id}>
              <td>{r.leaveTypeName || r.leaveTypeId}</td>
              <td>{r.startDate}</td>
              <td>{r.endDate}</td>
              <td style={{ maxWidth: 200, wordBreak: 'break-word' }}>{r.reason}</td>
              <td><StatusBadge status={r.status} /></td>
              <td style={{ color: '#666', fontStyle: r.managerComment ? 'normal' : 'italic' }}>
                {r.managerComment || '—'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

export default function EmployeePage() {
  const { user } = useUser()
  const [tab, setTab] = useState<Tab>('dashboard')
  const [leaveTypes, setLeaveTypes] = useState<LeaveType[]>([])
  const [balances, setBalances] = useState<LeaveBalance[]>([])
  const [requests, setRequests] = useState<LeaveRequest[]>([])
  const [balanceLoading, setBalanceLoading] = useState(true)
  const [requestsLoading, setRequestsLoading] = useState(true)

  const fetchBalance = useCallback(async () => {
    setBalanceLoading(true)
    try {
      const data = await getLeaveBalance(user!.id)
      setBalances(data)
    } catch {
      toast.error('Failed to load leave balances')
    } finally {
      setBalanceLoading(false)
    }
  }, [user])

  const fetchRequests = useCallback(async () => {
    setRequestsLoading(true)
    try {
      const data = await getLeaveRequests({ employeeId: user!.id })
      setRequests(data)
    } catch {
      toast.error('Failed to load leave requests')
    } finally {
      setRequestsLoading(false)
    }
  }, [user])

  useEffect(() => {
    getLeaveTypes()
      .then(setLeaveTypes)
      .catch(() => toast.error('Failed to load leave types'))
    fetchBalance()
    fetchRequests()
  }, [fetchBalance, fetchRequests])

  function handleSubmitSuccess() {
    fetchRequests()
    fetchBalance()
    setTab('history')
  }

  return (
    <div>
      <div className="tab-nav">
        <button className={`tab-btn ${tab === 'dashboard' ? 'active' : ''}`} onClick={() => setTab('dashboard')}>
          Dashboard
        </button>
        <button className={`tab-btn ${tab === 'submit' ? 'active' : ''}`} onClick={() => setTab('submit')}>
          Submit Request
        </button>
        <button className={`tab-btn ${tab === 'history' ? 'active' : ''}`} onClick={() => setTab('history')}>
          My Requests
        </button>
      </div>

      {tab === 'dashboard' && (
        <div className="card">
          <h3 className="card-title">Leave Balance — {user?.name}</h3>
          <BalanceDashboard balances={balances} loading={balanceLoading} />
        </div>
      )}

      {tab === 'submit' && (
        <div className="card">
          <h3 className="card-title">Submit Leave Request</h3>
          <SubmitForm leaveTypes={leaveTypes} onSuccess={handleSubmitSuccess} />
        </div>
      )}

      {tab === 'history' && (
        <div className="card">
          <h3 className="card-title">My Leave Requests</h3>
          <LeaveHistory requests={requests} loading={requestsLoading} />
        </div>
      )}
    </div>
  )
}
