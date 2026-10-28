import { useEffect, useState } from 'react'
import { deskRequest, apiBase, labels, words, minutes, date, rolesOf, Field, Select } from './desk-api'
import { DeskAdministration, CMDB } from './desk-admin'
import './desk.css'
import FAQ from './FAQ'

const routes = {
  nuevos: ['solicitado','analisis','pendiente_aprobacion','aprobado','implementacion','validacion','resuelto','cerrado'],
  cambios: ['solicitado','evaluacion','pendiente_aprobacion','aprobado','planificado','ejecucion','validacion','resuelto','cerrado'],
  incidentes: ['registrado','clasificado','diagnostico','atencion','recuperado','resuelto','cerrado'],
  eventos: ['registrado','clasificado','evaluado','respuesta','resuelto','cerrado'],
  problemas: ['registrado','investigacion','causa_identificada','solucion_propuesta','solucion_aplicada','resuelto','cerrado'],
  accesos: ['solicitado','validacion','pendiente_aprobacion','aprobado','provision_o_revocacion','verificacion','resuelto','cerrado'],
  servicios: ['solicitado','validacion','asignado','cumplimiento','confirmacion','resuelto','cerrado'],
}
export const processFields = {
  nuevos: { need:'Necesidad', feasibility:'Viabilidad', proposed_service_name:'Nombre del servicio propuesto', implementation_evidence:'Evidencia de implementación' },
  cambios: { impact_assessment:'Evaluación de impacto', risk:'Riesgo', test_plan:'Plan de pruebas', rollback_plan:'Plan de reversión', planned_at:'Fecha y plan de ejecución', execution_result:'Resultado de ejecución', validation_result:'Resultado de validación' },
  incidentes: { diagnosis:'Diagnóstico', recovery:'Recuperación', solution:'Solución aplicada', classification_reason:'Justificación de clasificación', redundancy_reason:'Justificación de redundancia' },
  eventos: { evaluation:'Evaluación', response:'Respuesta operativa' },
  problemas: { root_cause:'Causa raíz', workaround:'Solución temporal', definitive_solution:'Solución definitiva', exit_reason:'Justificación de salida' },
  accesos: { resource:'Recurso', permission:'Permiso solicitado', operation:'Acción grant o revoke', provision_evidence:'Evidencia de provisión o revocación', verification:'Verificación' },
  servicios: { request:'Solicitud concreta', fulfillment_evidence:'Evidencia de cumplimiento', confirmation:'Confirmación' },
}

export default function ServiceDesk({ session }) {
  const roles = rolesOf(session.user), readonly = roles.includes('consulta'), admin = roles.includes('admin') && !readonly
  const [view, setView] = useState('portal')
  const [context, setContext] = useState({ queues: [], services: [], users: [], processes: [] })
  const [error, setError] = useState(''), [ticket, setTicket] = useState(null)
  const [revision, setRevision] = useState(0)
  const api = (p, o) => deskRequest(p, session.token, o)
  async function loadContext() {
    try {
      const [queues, services, users, processes] = await Promise.all(['/api/desk/intake-queues','/api/catalog/services?limit=100','/api/users','/api/desk/processes'].map(p => api(p)))
      setContext({ queues, services, users, processes })
    } catch (e) { setError(e.message) }
  }
  useEffect(() => { loadContext() }, [session.token, revision])
  async function open(id) { try { setError(''); setTicket(await api(`/api/desk/tickets/${id}`)) } catch (e) { setError(e.message) } }
  return <main className="desk-shell">
    <div className="desk-title"><div><p className="eyebrow">SERVICE DESK</p><h2>Un solo lugar para dar seguimiento</h2><p className="muted">Solicitudes, atención y configuración de tus servicios.</p></div><span className="count-badge">{readonly ? 'Solo lectura' : roles.map(words).join(' · ')}</span></div>
    <nav className="desk-nav" aria-label="Service Desk">{[['portal','Mis solicitudes'],['faq','Centro de ayuda'], ...((roles.some(r => ['admin','agente','aprobador','consulta'].includes(r))) ? [['console','Consola de atención']] : []), ['indicators','Indicadores'], ['cmdb','CMDB'], ...(admin ? [['queues','Colas y permisos']] : [])].map(([key,label]) => <button key={key} className={view === key ? 'active' : 'secondary'} onClick={() => { setView(key); setTicket(null); setError('') }}>{label}</button>)}</nav>
    {error && <p role="alert" className="error-banner">{error}</p>}
    {ticket ? <Ticket key={ticket.id} initial={ticket} api={api} token={session.token} context={context} user={session.user} onBack={() => { setTicket(null); setRevision(x => x + 1) }} /> : <>
      {['portal','console'].includes(view) && <TicketList key={view} api={api} context={context} portal={view === 'portal'} readonly={readonly} user={session.user} onOpen={open} />}
      {view === 'indicators' && <Indicators api={api} context={context} onOpen={open} />}
      {view === 'queues' && admin && <DeskAdministration api={api} context={context} reload={() => setRevision(x => x + 1)} />}
      {view === 'cmdb' && <CMDB api={api} admin={admin} context={context} onTicket={open} />}
      {view === 'faq' && <FAQ api={api} admin={admin} services={context.services} />}
    </>}
  </main>
}

function TicketList({ api, context, portal, readonly, user, onOpen }) {
  const [items, setItems] = useState([]), [offset, setOffset] = useState(0), [error, setError] = useState(''), [creating, setCreating] = useState(false)
  const [filters, setFilters] = useState({ q:'', process:'', state:'', queue:'', priority:'', assignee:'', service:'', from:'', to:'' })
  const [authorizedQueues, setQueues] = useState([])
  const [notifications,setNotifications] = useState([])
  const limit = 20
  const set = (key, value) => { setOffset(0); setFilters(f => ({ ...f, [key]: value })) }
  async function load() {
    try { setError(''); const params = new URLSearchParams({ ...filters, limit, offset }); setItems(await api(`${portal ? '/api/desk/my-tickets' : '/api/desk/tickets'}?${params}`)) } catch (e) { setError(e.message) }
  }
  useEffect(() => { load() }, [JSON.stringify(filters), offset])
  useEffect(() => { api('/api/desk/queues').then(setQueues).catch(e => setError(e.message)) }, [])
  useEffect(() => { if(portal)api('/api/desk/notifications').then(setNotifications).catch(e=>setError(e.message)) }, [portal])
  const states = [...new Set(context.processes.filter(p => !filters.process || p.code === filters.process).flatMap(p => p.states || []))]
  return <section>
    <div className="section-heading"><div><h2>{portal ? 'Mis solicitudes' : 'Bandeja de atención'}</h2><p className="muted">{portal ? 'Consulta tus tickets y sus respuestas.' : 'Orden: prioridad, antigüedad e identificador. Solo colas autorizadas.'}</p></div>{!readonly && <button onClick={() => setCreating(!creating)}>{creating ? 'Cancelar' : 'Crear ticket'}</button>}</div>
    {creating && <CreateTicket api={api} context={context} user={user} onCreated={t => { setCreating(false); onOpen(t.id) }} />}
    {portal && notifications.length>0 && <details className="card form-card"><summary>Notificaciones verificables ({notifications.length})</summary>{notifications.map(n=><p key={n.id}>{n.message} · {date(n.created_at)}</p>)}</details>}
    <div className="desk-filters form-card"><Field label="Buscar por número o título" value={filters.q} onChange={e => set('q',e.target.value)} />
      {!portal && <><Select label="Proceso" value={filters.process} onChange={e => set('process',e.target.value)} items={Object.entries(labels).map(([code,name]) => ({code,name}))} empty="Todos" /><Select label="Cola" value={filters.queue} onChange={e => set('queue',e.target.value)} items={authorizedQueues} empty="Todas" /></>}
      <Select label="Estado" value={filters.state} onChange={e => set('state',e.target.value)} items={states.map(code => ({code,name:words(code)}))} empty="Todos" />
      {!portal && <><Select label="Prioridad" value={filters.priority} onChange={e => set('priority',e.target.value)} items={[{id:1,name:'Alta'},{id:2,name:'Media'},{id:3,name:'Baja'}]} empty="Todas" /><Select label="Responsable" value={filters.assignee} onChange={e => set('assignee',e.target.value)} items={context.users.map(u => ({id:u.id,name:u.full_name}))} empty="Todos" /><Select label="Servicio" value={filters.service} onChange={e => set('service',e.target.value)} enumerate items={context.services} empty="Todos" /><Field label="Desde" type="date" value={filters.from} onChange={e => set('from',e.target.value)} /><Field label="Hasta" type="date" value={filters.to} onChange={e => set('to',e.target.value)} /></>}
    </div>
    {error && <p role="alert" className="error">{error}</p>}
    <div className="table-card"><table><thead><tr><th>Ticket</th><th>Proceso</th><th>Estado</th><th>Prioridad</th><th>Responsable</th><th>Creado</th></tr></thead><tbody>{items.map(t => <tr key={t.id}><td><button className="desk-link" onClick={() => onOpen(t.id)}>{t.number}</button><div>{t.title}</div></td><td>{labels[t.process_code]}</td><td><span className="state-pill">{words(t.state)}</span></td><td>{['','Alta','Media','Baja'][t.priority]}</td><td>{context.users.find(u => u.id === t.assignee_id)?.full_name || 'Sin asignar'}</td><td>{date(t.created_at)}</td></tr>)}</tbody></table>{!items.length && <p className="desk-empty">No hay tickets para estos filtros.</p>}</div>
    <div className="desk-pagination"><button className="secondary" disabled={!offset} onClick={() => setOffset(Math.max(0,offset-limit))}>Anterior</button><span>Página {offset/limit+1}</span><button className="secondary" disabled={items.length < limit} onClick={() => setOffset(offset+limit)}>Siguiente</button></div>
  </section>
}

export function DetailsFields({ process, details, setDetails }) {
  const change = (key,value) => setDetails({ ...details, [key]:value })
  return <div className="desk-form-grid">
    {['eventos','incidentes'].includes(process) && <Select label="Categoría" value={details.category} onChange={e => change('category',e.target.value)} required items={Array.from({length:process === 'eventos'?4:7},(_,i) => ({code:`${process === 'eventos'?'E':'I'}${i}`,name:process === 'eventos' ? ['Programado','Sin impacto','Impacto al servicio','Seguridad'][i] : 'Categoría de incidente'}))} />}
    {process === 'incidentes' && <><Field label="Afecta seguridad"><select value={String(details.security ?? false)} onChange={e => change('security',e.target.value === 'true')}><option value="false">No</option><option value="true">Sí</option></select></Field><Field label="Interrupción o degradación"><select value={String(details.interrupted ?? false)} onChange={e => change('interrupted',e.target.value === 'true')}><option value="false">No</option><option value="true">Sí</option></select></Field><Field label="Redundancia disponible"><select value={details.redundancy == null ? '' : String(details.redundancy)} onChange={e => change('redundancy',e.target.value === '' ? null : e.target.value === 'true')}><option value="">Desconocida</option><option value="true">Sí</option><option value="false">No</option></select></Field>{['estimated_minutes','actual_minutes'].map(key => <Field key={key} label={key === 'estimated_minutes'?'Duración estimada (min)':'Duración real (min)'} type="number" min="0" step="any" value={details[key] ?? ''} onChange={e => change(key,e.target.value === '' ? null : Number(e.target.value))} />)}</>}
    {Object.entries(processFields[process] || {}).map(([key,label]) => <Field key={key} label={label}><textarea rows="2" value={details[key] ?? ''} onChange={e => change(key,e.target.value)} /></Field>)}
    {process === 'accesos' && <Field label="ID del usuario beneficiario" type="number" min="1" value={details.beneficiary_id ?? ''} onChange={e => change('beneficiary_id',e.target.value === '' ? null : Number(e.target.value))} />}
    {process === 'problemas' && <Field label="Error conocido"><input type="checkbox" checked={details.known_error || false} onChange={e => change('known_error',e.target.checked)} /></Field>}
  </div>
}

function CreateTicket({ api, context, user, onCreated }) {
  const [process,setProcess] = useState('servicios'), [details,setDetails] = useState({}), [cis,setCis] = useState([]), [search,setSearch] = useState(''), [selected,setSelected] = useState([]), [error,setError] = useState(''), [busy,setBusy] = useState(false)
  useEffect(() => { api(`/api/cmdb/items?limit=20&q=${encodeURIComponent(search)}`).then(setCis).catch(e => setError(e.message)) }, [search])
  async function submit(e) {
    e.preventDefault(); const f = new FormData(e.currentTarget); setBusy(true); setError('')
    try { const d = process === 'incidentes' ? { security:false,interrupted:false,redundancy:null,...details } : details
      const t = await api('/api/desk/tickets',{method:'POST',body:JSON.stringify({title:f.get('title'),description:f.get('description'),requester_id:user.id,service_level2_id:Number(f.get('service')),process_code:process,queue_id:Number(f.get('queue')),priority:Number(f.get('priority')),impact:f.get('impact'),urgency:f.get('urgency'),channel:'web',no_ci_reason:f.get('no_ci_reason'),ci_ids:selected.map(c => c.id),details:d})}); onCreated(t)
    } catch(e) { setError(e.message) } finally { setBusy(false) }
  }
  return <form className="card form-card desk-create" onSubmit={submit}><h3>Nueva solicitud</h3><div className="desk-form-grid">
    <Select label="Proceso" value={process} onChange={e => {setProcess(e.target.value);setDetails({})}} required items={Object.entries(labels).map(([code,name]) => ({code,name}))} />
    <Select key={process} label="Cola de recepción" name="queue" required items={context.queues.filter(q => q.process_code === process)} />
    <Field label="Título" name="title" required maxLength={200} /><Select label="Servicio nivel 2" name="service" required enumerate items={context.services.filter(s => s.is_active)} />
    <Field label="Descripción"><textarea name="description" required rows="3" /></Field>
    <Select label="Prioridad" name="priority" required items={[{id:1,name:'Alta'},{id:2,name:'Media'},{id:3,name:'Baja'}]} />
    {['impact','urgency'].map(k => <Select key={k} label={k === 'impact'?'Impacto':'Urgencia'} name={k} required items={['alto','medio','bajo'].map(code => ({code,name:code}))} />)}
    <Field label="Buscar CI afectados" value={search} onChange={e => setSearch(e.target.value)} /><Field label="Agregar CI"><select value="" onChange={e => {const c=cis.find(c => c.id === Number(e.target.value));if(c && !selected.some(x => x.id === c.id))setSelected([...selected,c])}}><option value="">Opcional</option>{cis.filter(c => c.is_active).map(c => <option value={c.id} key={c.id}>{c.code} · {c.name}</option>)}</select></Field>
    <div className="desk-chips">{selected.map(c => <button type="button" key={c.id} className="secondary" onClick={() => setSelected(selected.filter(x => x.id !== c.id))}>{c.code} ×</button>)}</div>
    {!selected.length && <Field label="Justificación sin CI" name="no_ci_reason" required />}
  </div><h4>Información de {labels[process]}</h4><DetailsFields process={process} details={details} setDetails={setDetails} />{error && <p role="alert" className="error">{error}</p>}<button disabled={busy}>{busy?'Creando…':'Enviar solicitud'}</button></form>
}

function Ticket({ initial, api, token, context, user, onBack }) {
  const [ticket,setTicket] = useState(initial), [details,setDetails] = useState(initial.details || {}), [attachments,setAttachments] = useState([]), [metrics,setMetrics] = useState(null), [error,setError] = useState(''), [notice,setNotice] = useState(''), [busy,setBusy] = useState(false), [action,setAction] = useState(initial.permissions?.handle?'transition':initial.permissions?.approve?'approve':'close')
  const p = ticket.permissions || {}, own = ticket.requester_id === user.id
  const path = `/api/desk/tickets/${ticket.id}`
  async function refresh() { const [t,a,m] = await Promise.all([api(path),api(`${path}/attachments`),api(`${path}/indicators`)]);setTicket(t);setDetails(t.details);setAttachments(a);setMetrics(m.items[0]) }
  useEffect(() => { refresh().catch(e => setError(e.message)) }, [initial.id])
  async function run(task) { setBusy(true);setError('');setNotice('');try {await task();await refresh();setNotice('Operación completada')}catch(e){setError(e.message);await refresh().catch(()=>{})}finally{setBusy(false)} }
  async function act(e) {e.preventDefault();const f=new FormData(e.currentTarget);await run(()=>api(`${path}/actions`,{method:'POST',body:JSON.stringify({action,version:ticket.version,to:f.get('to')||'',reason:f.get('reason')||'',resolution:f.get('resolution')||'',queue_id:Number(f.get('queue_id')||0),assignee_id:f.get('assignee_id')?Number(f.get('assignee_id')):null,incident_id:Number(f.get('incident_id')||0),details:action==='details'?details:action==='link_service'?{service_id:Number(f.get('result_service_id'))}:undefined})}))}
  async function download(a) { await run(async()=>{const res=await fetch(`${apiBase}${path}/attachments/${a.id}/download`,{headers:{Authorization:`Bearer ${token}`}});if(!res.ok){const b=await res.json();throw new Error(b.error)}const url=URL.createObjectURL(await res.blob());const link=document.createElement('a');link.href=url;link.download=a.original_name;link.click();setTimeout(()=>URL.revokeObjectURL(url),1000)}) }
  const allActions = [...(p.handle ? [['claim','Tomar ticket'],['assign','Asignar agente'],['transition','Cambiar estado'],['details','Editar información'],['pause','Pausar'],['resume','Reanudar'],['reevaluate','Reevaluar aprobación'],['process_event','Procesar evento'],['link_service','Vincular servicio resultado']] : []), ...(p.transfer ? [['transfer','Transferir'],['escalate','Escalar']] : []), ...(p.approve ? [['approve','Aprobar'],['reject','Rechazar']] : []), ...((p.handle || own) && p.write ? [['close','Confirmar cierre'],['reopen','Reabrir']] : [])]
  const route = routes[ticket.process_code] || [], next = route[route.indexOf(ticket.state)+1]
  const availableStates = [...new Set([next,...(ticket.process_code === 'eventos' && ticket.state === 'evaluado' ? ['resuelto']:[]),...(ticket.process_code === 'incidentes' && ticket.state === 'escalado'?['diagnostico']:[]),...(ticket.process_code === 'cambios' && ticket.state === 'validacion'?['planificado']:[])].filter(x => x && !['aprobado','rechazado','cerrado'].includes(x)))]
  const actions = allActions.filter(([a]) => {
    if(ticket.state==='rechazado') return false
    if(ticket.state==='cerrado') return a==='reopen'
    if(ticket.state==='resuelto') return ['close','reopen'].includes(a)
    if(ticket.paused) return a==='resume'
    if(['close','reopen','resume'].includes(a)) return false
    if(a==='claim') return ticket.assignee_id==null
    if(['approve','reject'].includes(a)) return ticket.state==='pendiente_aprobacion'
    if(a==='transition') return availableStates.length>0
    if(a==='process_event') return ticket.process_code==='eventos'&&ticket.state==='evaluado'
    if(a==='link_service') return ticket.process_code==='nuevos'&&ticket.state==='implementacion'
    if(a==='reevaluate') return ['nuevos','cambios','accesos'].includes(ticket.process_code)
    if(a==='escalate') return ticket.process_code==='incidentes'&&ticket.state==='atencion'
    return true
  })
  useEffect(()=>{if(!actions.some(([a])=>a===action)&&actions.length)setAction(actions[0][0])},[actions.map(([a])=>a).join(','),action])
  return <section><button className="secondary" onClick={onBack}>← Volver a la bandeja</button><div className="desk-ticket-heading"><div><p className="eyebrow">{ticket.number} · {labels[ticket.process_code]}</p><h2>{ticket.title}</h2></div><span className="state-pill">{words(ticket.state)}{ticket.paused?' · EN PAUSA':''}</span></div>
    {error && <p role="alert" className="error">{error}</p>}{notice && <p role="status" className="notice">{notice}</p>}
    <div className="desk-detail-grid"><div>
      <article className="card form-card"><h3>Detalle de la solicitud</h3><p className="desk-prose">{ticket.description}</p><dl className="desk-facts"><dt>Servicio</dt><dd>{context.services.find(s=>s.id===ticket.service_level2_id)?.name || ticket.service_level2_id}</dd><dt>Nivel 1 derivado</dt><dd>{ticket.service_level1_id}</dd><dt>Solicitante</dt><dd>{context.users.find(u=>u.id===ticket.requester_id)?.full_name}</dd><dt>Cola</dt><dd>{context.queues.find(q=>q.id===ticket.queue_id)?.name}</dd><dt>Responsable</dt><dd>{context.users.find(u=>u.id===ticket.assignee_id)?.full_name || 'Sin asignar'}</dd><dt>Prioridad</dt><dd>{['','Alta','Media','Baja'][ticket.priority]}</dd><dt>CI asociados</dt><dd>{ticket.ci?.map(c=>c.ci_id).join(', ')||ticket.no_ci_reason}</dd><dt>Creación</dt><dd>{date(ticket.created_at)}</dd><dt>Resolución</dt><dd>{ticket.resolution || 'Pendiente'}</dd></dl><details><summary>Información del proceso</summary>{Object.entries(ticket.details||{}).map(([k,v])=><p key={k}><strong>{processFields[ticket.process_code]?.[k] || words(k)}: </strong>{typeof v==='object'?JSON.stringify(v):String(v)}</p>)}</details></article>
      <article className="card form-card"><h3>Conversación</h3>{ticket.comments?.length ? ticket.comments.map(c=><div key={c.id} className={`desk-comment ${c.visibility==='internal'?'internal':''}`}><small>{context.users.find(u=>u.id===c.author_id)?.full_name} · {date(c.created_at)} · {c.visibility==='internal'?'Nota interna':'Público'}</small><p className="desk-prose">{c.body}</p></div>):<p className="muted">Todavía no hay comentarios.</p>}
      {p.write && <form onSubmit={async e=>{e.preventDefault();const form=e.currentTarget,f=new FormData(form);await run(async()=>{await api(`${path}/comments`,{method:'POST',body:JSON.stringify({body:f.get('body'),visibility:f.get('visibility')})});form.reset()})}}><Field label="Mensaje"><textarea name="body" required rows="3" /></Field><Field label="Visibilidad"><select name="visibility"><option value="public">Visible al solicitante</option>{p.internal&&<option value="internal">Nota interna</option>}</select></Field><button disabled={busy}>Enviar comentario</button></form>}</article>
      <article className="card form-card"><h3>Archivos de evidencia</h3><p className="muted">PDF, PNG, JPEG o TXT, hasta 10 MiB. Solo lectores autorizados del ticket.</p>{attachments.map(a=><div className="desk-file" key={a.id}><div><strong>{a.original_name}</strong><small>{(a.byte_size/1024).toFixed(1)} KiB · {a.visibility==='internal'?'Interno':'Público'}</small></div><button className="secondary" disabled={busy} onClick={()=>download(a)}>Descargar</button></div>)}
      {p.write && (own || p.handle) && !['cerrado','rechazado'].includes(ticket.state) && <form onSubmit={async e=>{e.preventDefault();const form=e.currentTarget;const f=new FormData(form);const file=f.get('file');if(file.size>10*1024*1024){setError('El archivo supera 10 MiB');return}await run(async()=>{await api(`${path}/attachments`,{method:'POST',body:f});form.reset()})}}><Field label="Archivo"><input type="file" name="file" accept=".pdf,.png,.jpg,.jpeg,.txt" required /></Field><Field label="Visibilidad"><select name="visibility"><option value="public">Público del ticket</option>{p.internal&&<option value="internal">Interno</option>}</select></Field><button disabled={busy}>Subir evidencia</button></form>}</article>
      <article className="card form-card"><h3>Bitácora</h3><ol className="desk-timeline">{ticket.history?.map(h=><li key={h.id}><strong>{words(h.action)}</strong><small>{date(h.occurred_at)} · Actor {h.actor_id}</small>{h.reason&&<p>{h.reason}</p>}{p.internal&&<details><summary>Valores anteriores y nuevos</summary><pre>{JSON.stringify({antes:h.before_value,despues:h.after_value},null,2)}</pre></details>}</li>)}</ol></article>
    </div><aside>
      {p.write && actions.length>0 && <form className="card form-card" onSubmit={act}><h3>Acciones del ticket</h3><Field label="Acción"><select value={action} onChange={e=>setAction(e.target.value)}>{actions.map(([key,label])=><option key={key} value={key}>{label}</option>)}</select></Field>
        {action==='transition'&&<Select key={`${ticket.version}-state`} label="Siguiente estado" name="to" required items={availableStates.map(code=>({code,name:words(code)}))} />}
        {['assign','transfer','escalate'].includes(action)&&<Select label="Agente destino" name="assignee_id" required={action==='assign'} items={context.users.filter(u=>u.is_active&&rolesOf(u).some(r=>['admin','agente'].includes(r))).map(u=>({id:u.id,name:u.full_name}))} empty={action==='assign'?'Selecciona…':'Sin asignar'} />}
        {['transfer','escalate'].includes(action)&&<Select label="Cola destino" name="queue_id" required items={context.queues.filter(q=>q.process_code===ticket.process_code&&q.id!==ticket.queue_id)} />}
        {action==='link_service'&&<Select label="Servicio resultado" name="result_service_id" required enumerate items={context.services.filter(s=>s.is_active)} />}
        {action==='process_event'&&<Field label="ID de incidente existente (opcional)" type="number" min="1" name="incident_id" />}
        {action==='details'&&<DetailsFields process={ticket.process_code} details={details} setDetails={setDetails} />}
        <Field label="Motivo"><textarea name="reason" rows="2" required={['transfer','escalate','approve','reject','pause','close','reopen','reevaluate'].includes(action)} /></Field>
        {action==='transition'&&<Field label="Resolución (obligatoria al resolver)"><textarea name="resolution" rows="3" /></Field>}
        <p className="muted">Versión {ticket.version}. El servidor valida permisos, aprobación y transiciones.</p><button disabled={busy}>{busy?'Procesando…':'Aplicar acción'}</button></form>}
      {metrics&&<TicketMetrics metrics={metrics} />}
      <article className="card form-card"><h3>Tickets relacionados</h3>{ticket.links?.map((l,i)=><p key={i}>{l.source_id} → {l.target_id} · {words(l.kind)}</p>)}{p.handle&&p.write&&<form onSubmit={async e=>{e.preventDefault();const f=new FormData(e.currentTarget);await run(()=>api(`${path}/links`,{method:'POST',body:JSON.stringify({target_id:Number(f.get('target_id')),kind:f.get('kind')})}))}}><Field label="ID del ticket destino" type="number" name="target_id" min="1" required /><Field label="Relación"><select name="kind"><option value="relacionado">Relacionado</option><option value="incidente_problema">Incidente → problema</option><option value="problema_cambio">Problema → cambio</option></select></Field><button disabled={busy}>Vincular ticket</button></form>}</article>
    </aside></div></section>
}

function TicketMetrics({ metrics:m }) {return <article className="card form-card"><h3>Tiempos y objetivos</h3><p className="muted">Minutos calendario, descontando la unión de pausas.</p><dl className="desk-facts"><dt>Primera atención</dt><dd>{minutes(m.first_attention_minutes)} · {words(m.first_status)}</dd><dt>Antigüedad abierta</dt><dd>{m.open_age_minutes==null?'No aplica (terminal)':minutes(m.open_age_minutes)}</dd><dt>Reaperturas</dt><dd>{m.reopens}</dd></dl>{m.cycles.map(c=><div className="desk-metric-cycle" key={c.cycle}><strong>Ciclo {c.cycle}</strong><p>Resolución: {minutes(c.resolution_minutes)}</p><p>Activo: {minutes(c.active_minutes)} · Pausa: {minutes(c.paused_minutes)}</p><span className="state-pill">{words(c.status)}</span></div>)}<details><summary>Objetivos históricos por cola</summary>{m.objective_segments.map(s=><p key={s.id}>Cola {s.queue_id} · ciclo {s.cycle}<br/>Atención: {s.first_response_minutes==null?'Sin objetivo':minutes(s.first_response_minutes)}<br/>Resolución: {s.resolution_minutes==null?'Sin objetivo':minutes(s.resolution_minutes)}<br/>{words(s.resolution_status)} · {s.source==='historical_unknown'?'Histórico desconocido':'Snapshot al ingreso'}</p>)}</details></article>}

function Indicators({ api, context, onOpen }) {
  const [data,setData]=useState(null),[error,setError]=useState(''),[process,setProcess]=useState(''),[queue,setQueue]=useState('')
  async function load(){try{setError('');setData(await api(`/api/desk/indicators?${new URLSearchParams({process,queue})}`))}catch(e){setError(e.message)}}
  useEffect(()=>{load()},[process,queue])
  const avg = key => {const values=data.items.map(t=>t[key]).filter(x=>x!=null);return values.length?minutes(values.reduce((a,b)=>a+b,0)/values.length):'Sin datos'}
  return <section><div className="section-heading"><div><h2>Indicadores de atención</h2><p className="muted">Solo tickets autorizados. Objetivos por tramo, sin reiniciar la antigüedad al transferir.</p></div><button className="secondary" onClick={load}>Actualizar</button></div><div className="desk-filters form-card"><Select label="Proceso" value={process} onChange={e=>setProcess(e.target.value)} items={Object.entries(labels).map(([code,name])=>({code,name}))} empty="Todos" /><Select label="Cola" value={queue} onChange={e=>setQueue(e.target.value)} items={context.queues} empty="Todas" /></div>{error&&<p role="alert" className="error">{error}</p>}{data&&<><div className="desk-stats">{[['Tickets visibles',data.total],['Primera atención promedio',avg('first_attention_minutes')],['Antigüedad abierta promedio',avg('open_age_minutes')],['Reaperturas',data.items.reduce((n,t)=>n+t.reopens,0)]].map(([label,value])=><article className="card" key={label}><small>{label}</small><strong>{value}</strong></article>)}</div><div className="desk-summary-grid">{[['Primera atención',data.first_attention],['Resolución del ciclo actual',data.resolution],['Por proceso',data.by_process],['Por estado',data.by_state],['Por prioridad',data.by_priority],['Por cola',data.by_queue]].map(([label,values])=><article className="card form-card" key={label}><h3>{label}</h3>{Object.entries(values).map(([k,v])=><div key={k} className="desk-summary-row"><span>{label==='Por cola'?context.queues.find(q=>String(q.id)===k)?.name||k:label==='Por prioridad'?['','Alta','Media','Baja'][k]:labels[k]||words(k)}</span><strong>{v}</strong></div>)}</article>)}</div><p className="muted">Calculado: {date(data.calculated_at)}. Sin objetivo no significa cumplido. Los abiertos no se promedian como resoluciones de cero minutos.</p><div className="table-card"><table><thead><tr><th>Ticket</th><th>Primera atención</th><th>Resolución actual</th><th>Pausas actuales</th><th>Reaperturas</th></tr></thead><tbody>{data.items.map(t=>{const c=t.cycles.at(-1);return <tr key={t.ticket_id}><td><button className="desk-link" onClick={()=>onOpen(t.ticket_id)}>{t.number}</button></td><td>{minutes(t.first_attention_minutes)}<small>{words(t.first_status)}</small></td><td>{minutes(c?.resolution_minutes)}<small>{words(c?.status)}</small></td><td>{minutes(c?.paused_minutes)}</td><td>{t.reopens}</td></tr>})}</tbody></table></div></>}</section>
}
