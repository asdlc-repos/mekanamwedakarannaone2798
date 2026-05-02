import { useUser } from '../context/UserContext'

export default function NavBar() {
  const { user, setUser } = useUser()

  return (
    <nav className="app-nav">
      <h1>Leave Management</h1>
      {user && (
        <div className="nav-user">
          <span>{user.name}</span>
          <span className="badge badge-role">{user.role}</span>
          <button className="btn btn-logout btn-sm" onClick={() => setUser(null)}>
            Logout
          </button>
        </div>
      )}
    </nav>
  )
}
