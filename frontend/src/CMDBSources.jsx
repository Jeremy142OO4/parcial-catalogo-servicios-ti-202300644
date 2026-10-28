import {useEffect,useState} from 'react'

const columns={B:'Nombre del CI',C:'Tipo de CI',D:'Descripción',E:'Número de serie',F:'Versión',G:'Fecha de adquisición',H:'Estado actual',I:'Relaciones',J:'Padres/hijos',K:'Ubicación física',L:'Propietario/responsable',M:'Fecha del cambio',N:'Descripción del cambio',O:'Documentación relacionada',P:'Enlaces a incidentes/problemas',Q:'Nivel de seguridad',R:'Cumplimiento',S:'Estado de configuración',T:'Número de licencia',U:'Vencimiento'}

export default function CMDBSources({api}){
 const [rows,setRows]=useState([]),[error,setError]=useState('')
 useEffect(()=>{api('/api/cmdb/sources').then(setRows).catch(e=>setError(e.message))},[api])
 return <section><h3>Origen del Excel</h3><p className="muted">Valores originales, sin modificar el archivo ni crear referencias inexistentes. Los ejemplos permanecen pendientes de revisión antes de convertirlos en CI operativos.</p>{error&&<p className="error" role="alert">{error}</p>}{!rows.length&&!error&&<p>No hay filas cargadas. Ejecuta la preparación de demostración o el generador para registrar la fuente.</p>}{rows.map(r=><article className="card form-card desk-spacing" key={r.id}><h4>{r.source_file} · {r.sheet} · fila {r.source_row}</h4><p className="muted" style={{overflowWrap:'anywhere'}}>SHA-256: {r.sha256}</p><ul>{r.observations.map((note,i)=><li key={i}>{note}</li>)}</ul><div className="table-card"><table><thead><tr><th>Celda</th><th>Campo</th><th>Valor original</th></tr></thead><tbody>{Object.entries(columns).map(([col,name])=><tr key={col}><td>{col}{r.source_row}</td><td>{name}</td><td style={{whiteSpace:'pre-wrap',overflowWrap:'anywhere'}}>{r.raw_values[`${col}${r.source_row}`]||'Sin información en origen'}</td></tr>)}</tbody></table></div></article>)}</section>
}
