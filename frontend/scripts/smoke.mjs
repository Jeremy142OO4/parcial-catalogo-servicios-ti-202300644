import assert from 'node:assert/strict'
import React from 'react'
import { renderToString } from 'react-dom/server'
import { createServer } from 'vite'
import { serviceOrdinal } from '../src/service-numbering.js'

assert.equal(serviceOrdinal(0),1)
assert.equal(serviceOrdinal(11),12)
assert.equal(serviceOrdinal(0,2),13)
assert.equal(serviceOrdinal(9,4),46)
console.log('PASS numeración de servicios: 1, 12, 13 y 46 sin reinicio entre páginas')

const server = await createServer({ server: { middlewareMode: true } })
try {
  const { Select } = await server.ssrLoadModule('/src/desk-api.jsx')
  const options = renderToString(React.createElement(Select,{label:'Servicio',enumerate:true,items:[{id:42,code:'SE.01.01',name:'Uno'},{id:73,code:'SE.01.02',name:'Dos'}]})).replaceAll('<!-- -->','')
  assert.match(options,/1\. SE\.01\.01/)
  assert.match(options,/2\. SE\.01\.02/)
  assert.match(options,/value="42"/)
  assert.match(options,/value="73"/)
  const { default: Desk, DetailsFields, processFields } = await server.ssrLoadModule('/src/ServiceDesk.jsx')
  const render = roles => renderToString(React.createElement(Desk, { session: { token:'synthetic-not-a-session', user:{ id:1, roles, role:roles[0], full_name:'Usuario sintético' } } }))
  const requester = render(['solicitante'])
  assert.match(requester, /Mis solicitudes/)
  assert.doesNotMatch(requester, /Consola de atención/)
  assert.doesNotMatch(requester, /Colas y permisos/)
  const admin = render(['admin'])
  assert.match(admin, /Colas y permisos/)
  assert.match(admin, /Consola de atención/)
  const readonly = render(['admin','agente','consulta'])
  assert.match(readonly, /Solo lectura/)
  assert.doesNotMatch(readonly, /Crear ticket/)
  assert.doesNotMatch(readonly, /Colas y permisos/)
  for (const process of ['nuevos','cambios','incidentes','eventos','problemas','accesos','servicios']) {
    const html = renderToString(React.createElement(DetailsFields,{process,details:{},setDetails:()=>{}}))
    assert.ok(html.includes('label'), `Missing differentiated fields: ${process}`)
  }
  assert.ok('impact_assessment' in processFields.cambios && 'planned_at' in processFields.cambios)
  assert.ok('provision_evidence' in processFields.accesos && 'operation' in processFields.accesos)
  assert.ok('definitive_solution' in processFields.problemas)
  assert.ok('fulfillment_evidence' in processFields.servicios)
  console.log('PASS UI render: solicitante, agente/admin, consulta con roles elevados y siete formularios diferenciados')
  console.log('Alcance: renderizado de componentes, no interacción real en navegador ni integración HTTP')
} finally { await server.close() }
