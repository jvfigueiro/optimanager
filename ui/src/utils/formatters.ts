export function formatBRL(value: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(value || 0)
}

export function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  // Se for YYYY-MM-DD
  if (dateStr.includes('-')) {
    const [year, month, day] = dateStr.split('T')[0].split('-')
    if (year && month && day) {
      return `${day}/${month}/${year}`
    }
  }
  return dateStr
}
