export const apiBase = import.meta.env.VITE_API_URL || ''
export async function deskRequest(path, token, options = {}) {
  const headers = { ...options.headers, Authorization: `Bearer ${token}` }
  if (options.body && !(options.body instanceof FormData)) headers['Content-Type'] = 'application/json'
  const res = await fetch(`${apiBase}${path}`, { ...options, headers })
  if (res.status === 204) return null
  const body = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(body.error || `Error HTTP ${res.status}`)
  return body
}
export const labels = { nuevos: 'Nuevos servicios', cambios: 'Cambios', incidentes: 'Incidentes', eventos: 'Eventos', problemas: 'Problemas', accesos: 'Accesos', servicios: 'Servicios existentes' }
export const words = v => String(v ?? 'Desconocido').replaceAll('_', ' ')
export const minutes = v => v == null ? 'Pendiente' : `${Number(v).toFixed(1)} min`
export const date = v => v ? new Date(v).toLocaleString('es-GT') : '—'
export const rolesOf = u => u.roles?.length ? u.roles : [u.role]
export function Field({ label, children, ...props }) { return <label>{label}{children || <input {...props} />}</label> }
export function Select({ label, value, onChange, items, enumerate = false, empty = 'Selecciona…', ...props }) {
  return <Field label={label}><select {...(value !== undefined ? {value:value ?? ''} : {})} onChange={onChange} {...props}><option value="">{empty}</option>{items.map((x,index) => <option key={x.id ?? x.code} value={x.id ?? x.code}>{enumerate ? `${index+1}. ` : ''}{x.code ? `${x.code} · ` : ''}{x.name}</option>)}</select></Field>
}
