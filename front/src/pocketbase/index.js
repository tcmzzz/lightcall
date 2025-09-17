import PocketBase from 'pocketbase'

const domain = window.location.hostname
export const pb = new PocketBase(`https://${domain}`)
pb.autoCancellation(false)

pb.beforeSend = function (url, options) {
  options.headers = Object.assign({}, options.headers, {})

  return { url, options }
}

pb.afterSend = async function (response, data) {
  if (response.status < 200 || response.status > 299) {
    console.error('API Error', response.status, response.statusText, data)
  }

  // auth fail
  if (response.status == 403) {
    console.error('Auth fail, clear authStore!', response.status, response.statusText, data)
    pb.authStore.clear()
    location.reload()
  }

  return data
}
