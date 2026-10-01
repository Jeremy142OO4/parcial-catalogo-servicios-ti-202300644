import { StrictMode, useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'

const API_URL = import.meta.env.VITE_API_URL || ''

async function request(path, options = {}, token = '') {
  const headers = { 'Content-Type': 'application/json', ...(options.headers || {}) }
  if (token) headers.Authorization = `Bearer ${token}`
  const response = await fetch(`${API_URL}${path}`, { ...options, headers })
  if (response.status === 204) return null
  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.error || 'No se pudo completar la operación')
  return body
}

function Login({ onLogin }) {
  const [identifier, setIdentifier] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(event) {
    event.preventDefault()
    setError('')
    setLoading(true)
    try {
      const session = await request('/api/auth/login', { method: 'POST', body: JSON.stringify({ identifier, password }) })
      onLogin(session)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return <main className="login-shell">
    <form className="card login-card" onSubmit={submit}>
      <p className="eyebrow">Software Avanzado</p>
      <h1>Catálogo de servicios TI</h1>
      <p className="muted">Ingresa con tu usuario o correo local.</p>
      <label>Usuario o correo<input value={identifier} onChange={(e) => setIdentifier(e.target.value)} required /></label>
      <label>Contraseña<input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required /></label>
      {error && <p className="error">{error}</p>}
      <button disabled={loading}>{loading ? 'Ingresando…' : 'Iniciar sesión'}</button>
    </form>
  </main>
}

function App() {
  const [session, setSession] = useState(() => JSON.parse(localStorage.getItem('catalogo_session') || 'null'))

  function login(next) {
    localStorage.setItem('catalogo_session', JSON.stringify(next))
    setSession(next)
  }

  async function logout() {
    if (session?.token) await request('/api/auth/logout', { method: 'POST' }, session.token).catch(() => {})
    localStorage.removeItem('catalogo_session')
    setSession(null)
  }

  if (!session) return <Login onLogin={login} />
  return <Dashboard session={session} onLogout={logout} />
}

function Dashboard({ session, onLogout }) {
  const [tab, setTab] = useState('services')
  const [services, setServices] = useState([])
  const [units, setUnits] = useState([])
  const [users, setUsers] = useState([])
  const [query, setQuery] = useState('')
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      setError('')
      const [nextServices, nextUnits, nextUsers] = await Promise.all([
        request(`/api/catalog/services?q=${encodeURIComponent(query)}`, {}, session.token),
        request('/api/organization/units', {}, session.token),
        request('/api/users', {}, session.token),
      ])
      setServices(nextServices)
      setUnits(nextUnits)
      setUsers(nextUsers)
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => { load() }, [query])

  async function createUnit(event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const parent = form.get('parent_id')
    try {
      await request('/api/organization/units', { method: 'POST', body: JSON.stringify({ type: form.get('type'), parent_id: parent ? Number(parent) : null, code: form.get('code'), name: form.get('name') }) }, session.token)
      event.currentTarget.reset(); setNotice('Unidad creada correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function createUser(event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    try {
      await request('/api/users', { method: 'POST', body: JSON.stringify({ position_id: Number(form.get('position_id')), full_name: form.get('full_name'), username: form.get('username'), email: form.get('email'), password: form.get('password'), role: form.get('role') }) }, session.token)
      event.currentTarget.reset(); setNotice('Usuario creado correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function assignService(event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const responsible = form.get('responsible_user_id')
    try {
      await request('/api/catalog/assignments', { method: 'POST', body: JSON.stringify({ service_id: Number(form.get('service_id')), section_id: Number(form.get('section_id')), responsible_user_id: responsible ? Number(responsible) : null }) }, session.token)
      event.currentTarget.reset(); setNotice('Asignación guardada correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  return <div className="app-shell">
    <header className="topbar">
      <div><p className="eyebrow">CATÁLOGO TI</p><h1>Panel de gestión</h1></div>
      <div className="session"><span>{session.user.full_name} · {session.user.role}</span><button className="secondary" onClick={onLogout}>Cerrar sesión</button></div>
    </header>
    <nav className="tabs">
      <button className={tab === 'services' ? 'active' : ''} onClick={() => setTab('services')}>Servicios</button>
      <button className={tab === 'organization' ? 'active' : ''} onClick={() => setTab('organization')}>Organización</button>
      {session.user.role === 'admin' && <button className={tab === 'admin' ? 'active' : ''} onClick={() => setTab('admin')}>Administración</button>}
    </nav>
    {notice && <div className="notice">{notice}</div>}
    {error && <div className="error-banner">{error}</div>}
    {tab === 'services' && <Services services={services} query={query} setQuery={setQuery} />}
    {tab === 'organization' && <Organization units={units} />}
    {tab === 'admin' && <Admin units={units} users={users} services={services} onUnitSubmit={createUnit} onUserSubmit={createUser} onAssignmentSubmit={assignService} />}
  </div>
}

function Services({ services, query, setQuery }) {
  return <section className="content">
    <div className="section-heading"><div><p className="eyebrow">CONSULTA</p><h2>Servicios de nivel 2</h2></div><input className="search" placeholder="Buscar por código o nombre" value={query} onChange={(e) => setQuery(e.target.value)} /></div>
    <div className="service-grid">{services.map((service) => <article className="card service-card" key={service.id}>
      <div className="service-code">{service.code}</div><h3>{service.name || 'Sin nombre'}</h3><p className="muted">{service.level1Code} · {service.level1Name}</p>
      <dl><div><dt>Clase</dt><dd>{service.serviceClass || 'Desconocida'}</dd></div><div><dt>Criticidad</dt><dd>{service.criticality || 'Desconocida'}</dd></div><div><dt>Estado</dt><dd>{service.activeValue || 'Sin información'}</dd></div></dl>
      {service.reviewRequired && <span className="badge">Revisión requerida</span>}
      <p className="assignment">{service.sectionName ? `Responsable: ${service.sectionName}` : 'Sin sección asignada'}</p>
    </article>)}</div>
  </section>
}

function Organization({ units }) {
  return <section className="content"><div className="section-heading"><div><p className="eyebrow">JERARQUÍA</p><h2>Organización</h2></div><span className="muted">{units.length} unidades</span></div><div className="table-card"><table><thead><tr><th>Tipo</th><th>Código</th><th>Nombre</th><th>Padre</th><th>Estado</th></tr></thead><tbody>{units.map((unit) => <tr key={`${unit.type}-${unit.id}`}><td>{unit.type}</td><td><strong>{unit.code}</strong></td><td>{unit.name}</td><td>{unit.parent_type || '—'} #{unit.parent_id || ''}</td><td><span className={unit.is_active ? 'status active-status' : 'status'}>{unit.is_active ? 'Activo' : 'Inactivo'}</span></td></tr>)}</tbody></table></div></section>
}

function Admin({ units, users, services, onUnitSubmit, onUserSubmit, onAssignmentSubmit }) {
  const positions = units.filter((unit) => unit.type === 'position' && unit.is_active)
  const sections = units.filter((unit) => unit.type === 'section' && unit.is_active)
  return <section className="content admin-grid"><form className="card form-card" onSubmit={onUnitSubmit}><p className="eyebrow">MANTENIMIENTO</p><h2>Nueva unidad</h2><label>Tipo<select name="type" defaultValue="company"><option value="company">Empresa</option><option value="area">Área</option><option value="department">Departamento</option><option value="section">Sección</option><option value="position">Puesto</option></select></label><label>ID del padre (opcional)<input name="parent_id" type="number" /></label><label>Código<input name="code" required /></label><label>Nombre<input name="name" required /></label><button>Crear unidad</button></form><form className="card form-card" onSubmit={onUserSubmit}><p className="eyebrow">ACCESOS</p><h2>Nuevo usuario</h2><label>Puesto<select name="position_id" required><option value="">Selecciona…</option>{positions.map((position) => <option key={position.id} value={position.id}>{position.code} · {position.name}</option>)}</select></label><label>Nombre completo<input name="full_name" required /></label><label>Usuario<input name="username" required /></label><label>Correo<input name="email" type="email" required /></label><label>Contraseña<input name="password" type="password" minLength="8" required /></label><label>Rol<select name="role" defaultValue="consulta"><option value="consulta">Consulta</option><option value="admin">Administrador</option></select></label><button>Crear usuario</button></form><form className="card form-card" onSubmit={onAssignmentSubmit}><p className="eyebrow">RESPONSABILIDADES</p><h2>Asignar servicio</h2><label>Servicio<select name="service_id" required><option value="">Selecciona…</option>{services.map((service) => <option key={service.id} value={service.id}>{service.code} · {service.name}</option>)}</select></label><label>Sección<select name="section_id" required><option value="">Selecciona…</option>{sections.map((section) => <option key={section.id} value={section.id}>{section.code} · {section.name}</option>)}</select></label><label>Responsable opcional<select name="responsible_user_id"><option value="">Sin responsable</option>{users.map((user) => <option key={user.id} value={user.id}>{user.full_name}</option>)}</select></label><button>Guardar asignación</button></form><div className="card form-card"><p className="eyebrow">USUARIOS</p><h2>Usuarios registrados</h2>{users.map((user) => <div className="user-row" key={user.id}><span><strong>{user.full_name}</strong><small>{user.username} · {user.role}</small></span><span className={user.is_active ? 'active-status' : 'muted'}>{user.is_active ? 'Activo' : 'Inactivo'}</span></div>)}</div></section>
}

createRoot(document.getElementById('root')).render(<StrictMode><App /></StrictMode>)
