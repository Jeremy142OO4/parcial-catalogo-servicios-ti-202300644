// Presentation position only. Never replaces a service's ID or original code.
export function serviceOrdinal(index, page = 1, pageSize = 12) {
  return (page - 1) * pageSize + index + 1
}
