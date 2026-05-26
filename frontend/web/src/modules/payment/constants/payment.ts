export const paymentTypeOptions = [
  { label: '全部类型', value: '' },
  { label: '报名缴费', value: 'signup' },
  { label: '续费缴费', value: 'renewal' },
]

export const paymentStatusOptions = [
  { label: '全部状态', value: '' },
  { label: '已缴', value: 'paid' },
  { label: '待缴', value: 'pending' },
  { label: '部分已缴', value: 'partial' },
  { label: '已退款', value: 'refunded' },
]

export const paymentMethodOptions = [
  { label: '全部方式', value: '' },
  { label: '现金', value: 'cash' },
  { label: '微信', value: 'wechat' },
  { label: '支付宝', value: 'alipay' },
  { label: '银行转账', value: 'bank_transfer' },
  { label: '其他', value: 'other' },
]
