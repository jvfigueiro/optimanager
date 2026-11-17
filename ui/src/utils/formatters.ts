export function reaisToCentavos(reais: number): number {
  return Math.round((Number(reais) || 0) * 100)
}

export function centavosToReais(centavos: number): number {
  return (Number(centavos) || 0) / 100
}

export function formatCentavos(centavos: number): string {
  return new Intl.NumberFormat('pt-BR', {
    style: 'currency',
    currency: 'BRL',
  }).format(centavosToReais(centavos))
}

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

const PROBLEMA_PREFIX = '[Problema Informado]\n'
const OBS_SEPARATOR = '\n\n[Observações]\n'

export function splitObsAndProblema(rawObs?: string): { problema: string; obs: string } {
  const text = (rawObs || '').trim()
  if (!text) return { problema: '', obs: '' }

  if (text.startsWith('[Problema Informado]')) {
    const rest = text.slice('[Problema Informado]'.length).replace(/^\r?\n/, '')
    const sepIdx = rest.indexOf(OBS_SEPARATOR)
    if (sepIdx !== -1) {
      return {
        problema: rest.slice(0, sepIdx).trim(),
        obs: rest.slice(sepIdx + OBS_SEPARATOR.length).trim(),
      }
    }
    return { problema: rest.trim(), obs: '' }
  }

  return { problema: '', obs: text }
}

export function mergeObsAndProblema(problema?: string, obs?: string): string {
  const p = (problema || '').trim()
  const o = (obs || '').trim()
  if (p && o) {
    return `${PROBLEMA_PREFIX}${p}${OBS_SEPARATOR}${o}`
  }
  if (p) {
    return `${PROBLEMA_PREFIX}${p}`
  }
  return o
}


