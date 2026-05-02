import { useUser, DEMO_USERS } from '../context/UserContext'

export default function LoginPage() {
  const { setUser } = useUser()

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-logo">
          <h2>Leave Management</h2>
          <p>Select your profile to continue</p>
        </div>

        <div>
          {DEMO_USERS.map(u => (
            <button
              key={u.id}
              className="user-select-btn"
              onClick={() => setUser(u)}
            >
              <div>
                <div className="user-select-name">{u.name}</div>
                <div className="user-select-role">ID: {u.id}</div>
              </div>
              <span className="badge" style={{
                background: u.role === 'manager' ? '#e8f0fe' : '#e6f4ea',
                color: u.role === 'manager' ? '#1a73e8' : '#137333'
              }}>
                {u.role}
              </span>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
