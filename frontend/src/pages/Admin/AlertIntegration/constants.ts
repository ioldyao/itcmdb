// Webhook 类型映射常量
export const WEBHOOK_TYPE_MAP = {
  alertmanager: { text: 'Alertmanager', color: 'orange' },
  prometheus: { text: 'Prometheus', color: 'blue' },
  victoriametrics: { text: 'VictoriaMetrics', color: 'green' },
  custom: { text: '自定义', color: 'purple' },
  receiver: { text: '告警接收人', color: 'blue' },
} as const

// 接收人类型映射常量
export const RECEIVER_TYPE_MAP = {
  wechat: { text: '企业微信', color: 'green' },
  dingtalk: { text: '钉钉', color: 'blue' },
  feishu: { text: '飞书', color: 'cyan' },
  email: { text: '邮件', color: 'purple' },
  sms: { text: '短信', color: 'magenta' },
} as const

// 时间常量
export const ONE_HOUR_MS = 3600000
export const DEFAULT_PAGE_SIZE = 1000
