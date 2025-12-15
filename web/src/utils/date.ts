import dayjs from 'dayjs'

export const formatDate = (timestamp: number | string, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!timestamp) return '-'
  // If timestamp is a number and likely in seconds (10 digits), convert to milliseconds
  if (typeof timestamp === 'number' && timestamp.toString().length === 10) {
    return dayjs.unix(timestamp).format(format)
  }
  return dayjs(timestamp).format(format)
}