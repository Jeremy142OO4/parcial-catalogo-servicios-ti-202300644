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
  const [level1s, setLevel1s] = useState([])
  const [lookups, setLookups] = useState({ service_classes: [], criticalities: [], service_types: [] })
  const [units, setUnits] = useState([])
  const [users, setUsers] = useState([])
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState({ level1_id: '', is_active: 'true', service_class: '', criticality: '', service_type: '' })
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      setError('')
      const params = new URLSearchParams({ q: query, ...Object.fromEntries(Object.entries(filters).filter(([, value]) => value !== '')) })
      const [nextServices, nextUnits, nextUsers, nextLevel1s, nextLookups] = await Promise.all([
        request(`/api/catalog/services?${params}`, {}, session.token),
        request('/api/organization/units', {}, session.token),
        request('/api/users', {}, session.token),
        request('/api/catalog/level1', {}, session.token),
        request('/api/catalog/lookups', {}, session.token),
      ])
      setServices(nextServices)
      setUnits(nextUnits)
      setUsers(nextUsers)
      setLevel1s(nextLevel1s)
      setLookups(nextLookups)
    } catch (err) {
      setError(err.message)
    }
  }

  useEffect(() => { load() }, [query, filters.level1_id, filters.is_active, filters.service_class, filters.criticality, filters.service_type])

  async function createUnit(event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const parent = form.get('parent_id')
    try {
      await request('/api/organization/units', { method: 'POST', body: JSON.stringify({ type: form.get('type'), parent_id: parent ? Number(parent) : null, code: form.get('code'), name: form.get('name') }) }, session.token)
      event.currentTarget.reset(); setNotice('Unidad creada correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function updateUnit(unit, event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    try {
      await request(`/api/organization/units/${unit.type}/${unit.id}`, { method: 'PATCH', body: JSON.stringify({ code: form.get('code'), name: form.get('name') }) }, session.token)
      setNotice('Unidad actualizada correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function toggleUnit(unit) {
    try {
      await request(`/api/organization/units/${unit.type}/${unit.id}/active`, { method: 'PATCH', body: JSON.stringify({ is_active: !unit.is_active }) }, session.token)
      setNotice(`Unidad ${unit.is_active ? 'desactivada' : 'activada'}`); load()
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

  async function updateUser(user, event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    try {
      await request(`/api/users/${user.id}`, { method: 'PATCH', body: JSON.stringify({ position_id: Number(form.get('position_id')), full_name: form.get('full_name'), email: form.get('email'), role: form.get('role') }) }, session.token)
      setNotice('Usuario actualizado correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function toggleUser(user) {
    try {
      await request(`/api/users/${user.id}/active`, { method: 'PATCH', body: JSON.stringify({ is_active: !user.is_active }) }, session.token)
      setNotice(`Usuario ${user.is_active ? 'desactivado' : 'activado'}`); load()
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

  async function createLevel1(event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    try {
      await request('/api/catalog/level1', { method: 'POST', body: JSON.stringify({ code: form.get('code'), name: form.get('name') }) }, session.token)
      event.currentTarget.reset(); setNotice('Servicio nivel 1 creado correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function updateLevel1(level1, event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    try {
      await request(`/api/catalog/level1/${level1.id}`, { method: 'PATCH', body: JSON.stringify({ code: form.get('code'), name: form.get('name') }) }, session.token)
      setNotice('Servicio nivel 1 actualizado correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function createLevel2(event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const optional = (name) => form.get(name) || null
    const number = (name) => form.get(name) === '' ? null : Number(form.get(name))
    try {
      await request('/api/catalog/services', { method: 'POST', body: JSON.stringify({ level1_id: Number(form.get('level1_id')), code: form.get('code'), name: form.get('name'), service_class: optional('service_class'), criticality: optional('criticality'), service_type: optional('service_type'), minimum: number('minimum'), maximum: number('maximum') }) }, session.token)
      event.currentTarget.reset(); setNotice('Servicio nivel 2 creado correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function updateLevel2(service, event) {
    event.preventDefault()
    const form = new FormData(event.currentTarget)
    const optional = (name) => form.get(name) || null
    const number = (name) => form.get(name) === '' ? null : Number(form.get(name))
    try {
      await request(`/api/catalog/services/${service.id}`, { method: 'PATCH', body: JSON.stringify({ level1_id: Number(form.get('level1_id')), code: form.get('code'), name: form.get('name'), active_value: optional('active_value'), service_class: optional('service_class'), criticality: optional('criticality'), service_type: optional('service_type'), minimum: number('minimum'), maximum: number('maximum') }) }, session.token)
      setNotice('Servicio nivel 2 actualizado correctamente'); load()
    } catch (err) { setError(err.message) }
  }

  async function toggleService(service) {
    try {
      await request(`/api/catalog/services/${service.id}/active`, { method: 'PATCH', body: JSON.stringify({ is_active: !service.is_active }) }, session.token)
      setNotice(`Servicio ${service.is_active ? 'desactivado' : 'activado'}`); load()
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
    {tab === 'services' && <Services services={services} query={query} setQuery={setQuery} filters={filters} setFilters={setFilters} level1s={level1s} lookups={lookups} isAdmin={session.user.role === 'admin'} onToggleService={toggleService} />}
    {tab === 'organization' && <Organization units={units} />}
    {tab === 'admin' && <><Admin units={units} users={users} services={services} onUnitSubmit={createUnit} onUnitUpdate={updateUnit} onUnitToggle={toggleUnit} onUserSubmit={createUser} onUserUpdate={updateUser} onUserToggle={toggleUser} onAssignmentSubmit={assignService} /><CatalogAdmin level1s={level1s} services={services} lookups={lookups} onLevel1Submit={createLevel1} onLevel1Update={updateLevel1} onLevel2Submit={createLevel2} onLevel2Update={updateLevel2} onToggleService={toggleService} /></>}
  </div>
}

function Services({ services, query, setQuery, filters, setFilters, level1s, lookups, isAdmin, onToggleService }) {
  const setFilter = (name, value) => setFilters((current) => ({ ...current, [name]: value }))
  return <section className="content">
    <div className="section-heading"><div><p className="eyebrow">CONSULTA</p><h2>Servicios de nivel 2</h2></div><input className="search" placeholder="Buscar por código o nombre" value={query} onChange={(e) => setQuery(e.target.value)} /></div>
    <div className="filters"><select value={filters.level1_id} onChange={(e) => setFilter('level1_id', e.target.value)}><option value="">Todos los niveles 1</option>{level1s.map((level1) => <option key={level1.id} value={level1.id}>{level1.code} · {level1.name}</option>)}</select><select value={filters.service_class} onChange={(e) => setFilter('service_class', e.target.value)}><option value="">Todas las clases</option>{lookups.service_classes.map((value) => <option key={value}>{value}</option>)}</select><select value={filters.criticality} onChange={(e) => setFilter('criticality', e.target.value)}><option value="">Todas las criticidades</option>{lookups.criticalities.map((value) => <option key={value}>{value}</option>)}</select><select value={filters.service_type} onChange={(e) => setFilter('service_type', e.target.value)}><option value="">Todos los tipos</option>{lookups.service_types.map((value) => <option key={value}>{value}</option>)}</select><select value={filters.is_active} onChange={(e) => setFilter('is_active', e.target.value)}><option value="">Todos los estados</option><option value="true">Activos</option><option value="false">Inactivos</option></select></div>
    <div className="service-grid">{services.map((service) => <article className="card service-card" key={service.id}>
      <div className="service-code">{service.code}</div><h3>{service.name || 'Sin nombre'}</h3><p className="muted">{service.level1_code} · {service.level1_name}</p>
      <dl><div><dt>Clase</dt><dd>{service.service_class || 'Desconocida'}</dd></div><div><dt>Criticidad</dt><dd>{service.criticality || 'Desconocida'}</dd></div><div><dt>Estado</dt><dd>{service.is_active ? 'Activo' : 'Inactivo'}</dd></div></dl>
      {service.review_required && <span className="badge">Revisión requerida</span>}
      <p className="assignment">{service.section_name ? `Responsable: ${service.section_name}` : 'Sin sección asignada'}</p>
      {isAdmin && <button className="small-button" onClick={() => onToggleService(service)}>{service.is_active ? 'Desactivar' : 'Activar'}</button>}
    </article>)}</div>
  </section>
}

function Organization({ units }) {
  return <section className="content"><div className="section-heading"><div><p className="eyebrow">JERARQUÍA</p><h2>Organización</h2></div><span className="muted">{units.length} unidades</span></div><div className="table-card"><table><thead><tr><th>Tipo</th><th>Código</th><th>Nombre</th><th>Padre</th><th>Estado</th></tr></thead><tbody>{units.map((unit) => <tr key={`${unit.type}-${unit.id}`}><td>{unit.type}</td><td><strong>{unit.code}</strong></td><td>{unit.name}</td><td>{unit.parent_type || '—'} #{unit.parent_id || ''}</td><td><span className={unit.is_active ? 'status active-status' : 'status'}>{unit.is_active ? 'Activo' : 'Inactivo'}</span></td></tr>)}</tbody></table></div></section>
}

function Admin({ units, users, services, onUnitSubmit, onUnitUpdate, onUnitToggle, onUserSubmit, onUserUpdate, onUserToggle, onAssignmentSubmit }) {
  const positions = units.filter((unit) => unit.type === 'position' && unit.is_active)
  const sections = units.filter((unit) => unit.type === 'section' && unit.is_active)
  return <section className="content admin-grid"><form className="card form-card" onSubmit={onUnitSubmit}><p className="eyebrow">MANTENIMIENTO</p><h2>Nueva unidad</h2><label>Tipo<select name="type" defaultValue="company"><option value="company">Empresa</option><option value="area">Área</option><option value="department">Departamento</option><option value="section">Sección</option><option value="position">Puesto</option></select></label><label>ID del padre (opcional)<input name="parent_id" type="number" /></label><label>Código<input name="code" required /></label><label>Nombre<input name="name" required /></label><button>Crear unidad</button></form><form className="card form-card" onSubmit={onUserSubmit}><p className="eyebrow">ACCESOS</p><h2>Nuevo usuario</h2><label>Puesto<select name="position_id" required><option value="">Selecciona…</option>{positions.map((position) => <option key={position.id} value={position.id}>{position.code} · {position.name}</option>)}</select></label><label>Nombre completo<input name="full_name" required /></label><label>Usuario<input name="username" required /></label><label>Correo<input name="email" type="email" required /></label><label>Contraseña<input name="password" type="password" minLength="8" required /></label><label>Rol<select name="role" defaultValue="consulta"><option value="consulta">Consulta</option><option value="admin">Administrador</option></select></label><button>Crear usuario</button></form><form className="card form-card" onSubmit={onAssignmentSubmit}><p className="eyebrow">RESPONSABILIDADES</p><h2>Asignar servicio</h2><label>Servicio<select name="service_id" required><option value="">Selecciona…</option>{services.map((service) => <option key={service.id} value={service.id}>{service.code} · {service.name}</option>)}</select></label><label>Sección<select name="section_id" required><option value="">Selecciona…</option>{sections.map((section) => <option key={section.id} value={section.id}>{section.code} · {section.name}</option>)}</select></label><label>Responsable opcional<select name="responsible_user_id"><option value="">Sin responsable</option>{users.map((user) => <option key={user.id} value={user.id}>{user.full_name}</option>)}</select></label><button>Guardar asignación</button></form><UnitMaintenance units={units} onUpdate={onUnitUpdate} onToggle={onUnitToggle} /><UserMaintenance users={users} positions={positions} onUpdate={onUserUpdate} onToggle={onUserToggle} /></section>
}

function UnitMaintenance({ units, onUpdate, onToggle }) {
  const [selected, setSelected] = useState('')
  const key = selected || (units[0] ? `${units[0].type}:${units[0].id}` : '')
  const unit = units.find((item) => `${item.type}:${item.id}` === key)
  return <div className="card form-card"><p className="eyebrow">MANTENIMIENTO</p><h2>Editar unidad</h2><label>Unidad<select value={key} onChange={(e) => setSelected(e.target.value)}>{units.map((item) => <option key={`${item.type}:${item.id}`} value={`${item.type}:${item.id}`}>{item.type} · {item.code}</option>)}</select></label>{unit && <form onSubmit={(event) => onUpdate(unit, event)}><label>Código<input name="code" defaultValue={unit.code} required /></label><label>Nombre<input name="name" defaultValue={unit.name} required /></label><div className="button-row"><button>Guardar cambios</button><button type="button" className="secondary" onClick={() => onToggle(unit)}>{unit.is_active ? 'Desactivar' : 'Activar'}</button></div></form>}</div>
}

function UserMaintenance({ users, positions, onUpdate, onToggle }) {
  const [selected, setSelected] = useState('')
  const user = users.find((item) => String(item.id) === (selected || String(users[0]?.id)))
  return <div className="card form-card"><p className="eyebrow">ACCESOS</p><h2>Editar usuario</h2><label>Usuario<select value={user?.id || ''} onChange={(e) => setSelected(e.target.value)}>{users.map((item) => <option key={item.id} value={item.id}>{item.full_name} · {item.username}</option>)}</select></label>{user && <form onSubmit={(event) => onUpdate(user, event)}><label>Puesto<select name="position_id" defaultValue={user.position_id}>{positions.map((position) => <option key={position.id} value={position.id}>{position.code} · {position.name}</option>)}</select></label><label>Nombre completo<input name="full_name" defaultValue={user.full_name} required /></label><label>Correo<input name="email" type="email" defaultValue={user.email} required /></label><label>Rol<select name="role" defaultValue={user.role}><option value="consulta">Consulta</option><option value="admin">Administrador</option></select></label><div className="button-row"><button>Guardar cambios</button><button type="button" className="secondary" onClick={() => onToggle(user)}>{user.is_active ? 'Desactivar' : 'Activar'}</button></div></form>}</div>
}

function CatalogAdmin({ level1s, services, lookups, onLevel1Submit, onLevel1Update, onLevel2Submit, onLevel2Update, onToggleService }) {
  const [selectedLevel1, setSelectedLevel1] = useState('')
  const [selectedService, setSelectedService] = useState('')
  const level1 = level1s.find((item) => String(item.id) === (selectedLevel1 || String(level1s[0]?.id)))
  const service = services.find((item) => String(item.id) === (selectedService || String(services[0]?.id)))
  return <section className="content admin-grid"><form className="card form-card" onSubmit={onLevel1Submit}><p className="eyebrow">CATÁLOGO</p><h2>Nuevo nivel 1</h2><label>Código<input name="code" required /></label><label>Nombre canónico<input name="name" required /></label><button>Crear nivel 1</button></form><form className="card form-card" onSubmit={onLevel2Submit}><p className="eyebrow">CATÁLOGO</p><h2>Nuevo nivel 2</h2><label>Servicio nivel 1<select name="level1_id" required><option value="">Selecciona…</option>{level1s.map((item) => <option key={item.id} value={item.id}>{item.code} · {item.name}</option>)}</select></label><label>Código<input name="code" required placeholder="SE.99.1" /></label><label>Nombre<input name="name" required /></label><label>Clase<select name="service_class"><option value="">Desconocida</option>{lookups.service_classes.map((value) => <option key={value}>{value}</option>)}</select></label><label>Criticidad<select name="criticality"><option value="">Desconocida</option>{lookups.criticalities.map((value) => <option key={value}>{value}</option>)}</select></label><label>Tipo<select name="service_type"><option value="">Desconocido</option>{lookups.service_types.map((value) => <option key={value}>{value}</option>)}</select></label><div className="two-fields"><label>Mínimo<input name="minimum" type="number" step="any" /></label><label>Máximo<input name="maximum" type="number" step="any" /></label></div><button>Crear nivel 2</button></form><div className="card form-card"><p className="eyebrow">CATÁLOGO</p><h2>Editar nivel 1</h2><label>Servicio<select value={level1?.id || ''} onChange={(e) => setSelectedLevel1(e.target.value)}>{level1s.map((item) => <option key={item.id} value={item.id}>{item.code} · {item.name}</option>)}</select></label>{level1 && <form onSubmit={(event) => onLevel1Update(level1, event)}><label>Código<input name="code" defaultValue={level1.code} required /></label><label>Nombre canónico<input name="name" defaultValue={level1.name} required /></label><button>Guardar cambios</button></form>}</div><div className="card form-card"><p className="eyebrow">CATÁLOGO</p><h2>Editar nivel 2</h2><label>Servicio<select value={service?.id || ''} onChange={(e) => setSelectedService(e.target.value)}>{services.map((item) => <option key={item.id} value={item.id}>{item.code} · {item.name}</option>)}</select></label>{service && <form onSubmit={(event) => onLevel2Update(service, event)}><label>Nivel 1<select name="level1_id" defaultValue={service.level1_id} required>{level1s.map((item) => <option key={item.id} value={item.id}>{item.code} · {item.name}</option>)}</select></label><label>Código<input name="code" defaultValue={service.code} required /></label><label>Nombre<input name="name" defaultValue={service.name} required /></label><label>Activo<select name="active_value" defaultValue={service.active_value || ''}><option value="">Desconocido</option><option value="S">S</option><option value="N">N</option></select></label><label>Clase<select name="service_class" defaultValue={service.service_class || ''}><option value="">Desconocida</option>{lookups.service_classes.map((value) => <option key={value}>{value}</option>)}</select></label><label>Criticidad<select name="criticality" defaultValue={service.criticality || ''}><option value="">Desconocida</option>{lookups.criticalities.map((value) => <option key={value}>{value}</option>)}</select></label><label>Tipo<select name="service_type" defaultValue={service.service_type || ''}><option value="">Desconocido</option>{lookups.service_types.map((value) => <option key={value}>{value}</option>)}</select></label><div className="two-fields"><label>Mínimo<input name="minimum" type="number" step="any" defaultValue={service.minimum ?? ''} /></label><label>Máximo<input name="maximum" type="number" step="any" defaultValue={service.maximum ?? ''} /></label></div><div className="button-row"><button>Guardar cambios</button><button type="button" className="secondary" onClick={() => onToggleService(service)}>{service.is_active ? 'Desactivar' : 'Activar'}</button></div></form>}</div></section>
}

createRoot(document.getElementById('root')).render(<StrictMode><App /></StrictMode>)
