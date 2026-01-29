import { useMemo } from 'react'
import { ONE_HOUR_MS } from '../constants'

interface Webhook {
  enabled: boolean
  last_received?: string
  last_sent?: string
}

export function useInboundWebhookStats(webhooks: Webhook[]) {
  return useMemo(() => {
    const total = webhooks.length
    const enabled = webhooks.filter(w => w.enabled).length
    const activeInLastHour = webhooks.filter(
      w => w.last_received && new Date(w.last_received).getTime() > Date.now() - ONE_HOUR_MS
    ).length
    const withRecords = webhooks.filter(w => w.last_received).length

    return { total, enabled, activeInLastHour, withRecords }
  }, [webhooks])
}

export function useOutboundWebhookStats(webhooks: Webhook[]) {
  return useMemo(() => {
    const total = webhooks.length
    const enabled = webhooks.filter(w => w.enabled).length
    const activeInLastHour = webhooks.filter(
      w => w.last_sent && new Date(w.last_sent).getTime() > Date.now() - ONE_HOUR_MS
    ).length

    return { total, enabled, activeInLastHour }
  }, [webhooks])
}
