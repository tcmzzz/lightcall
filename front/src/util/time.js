import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn' // ES 2015

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

export const formatAbsoluteTime = (dateStr) => {
  return dayjs(dateStr).format('YYYY-MM-DD HH:mm')
}

export const formatRelativeTime = (dateStr) => {
  return dayjs(dateStr).fromNow()
}

export default dayjs
