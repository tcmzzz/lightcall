import JsSIP from 'jssip'
import { pb } from '@/pocketbase'
import { useConfigStore } from '@/stores/config'

//JsSIP.debug.disable('JsSIP:*')
JsSIP.debug.enable('JsSIP:*')

var Phone = {
  ua: null,
  status: ref(''),
  session: {
    status: ref(''),
    _session: null
  },
  bind: false,
  init: function () {
    if (this.ua !== null) {
      return
    }

    const userId = pb.authStore.record.id
    const domain = window.location.hostname
    const wssAddr = `wss:/${domain}/wscall/fs`
    const uri = `sip:${userId}@${domain}`

    const socket = new JsSIP.WebSocketInterface(wssAddr)
    const configuration = {
      sockets: [socket],
      register: false,
      uri: uri
      //register_expires: 5,
      //extra_headers: ['Foo: ABC', 'Bar: XYZ'],
      //password: '1234'
    }
    this.ua = new JsSIP.UA(configuration)
  },
  start: function () {
    return new Promise((resolve, reject) => {
      if (this.ua === null) {
        reject('ua not init')
      }
      if (this.ua.isConnected()) {
        //reject('ua is connected')
        resolve('connected')
      }
      if (!this.bind) {
        const ua = this.ua
        const status = this.status

        ListenPEvs.forEach((e) => {
          ua.on(e, function (data) {
            status.value = e
            console.debug('phone event', e, data)
            if (e === 'connected') {
              resolve('connected')
            }
            if (e === 'disconnected') {
              reject('disconnected')
            }
          })
        })

        this.bind = true
      }
      this.ua.start()
    })
  },
  stop: function () {
    this.hangup()

    if (this.ua != null) {
      this.ua.stop()
      this.ua.removeAllListeners()
    }
    this.ua = null
    this.bind = false
    this.status.value = ''
  },
  hangup: function () {
    if (this.session._session != null && this.session._session.isEstablished()) {
      this.session._session.terminate()
    }
    if (this.session._session != null) {
      this.session._session.removeAllListeners()
    }
    this.session._session = null
    //this.session.status.value = ''
  },
  call: function (taskId, activityId, endFunc) {
    return new Promise(async (resolve, reject) => {
      if (this.ua === null || !this.ua.isConnected()) {
        reject('not connected')
      }
      if (this.session._session !== null && !this.session._session.isEnded()) {
        reject('session not ended')
      }

      const iceServers = await useConfigStore().getIceServers()
      console.log('ice!!!', iceServers)

      const userId = pb.authStore.record.id
      const token = pb.authStore.token

      const options = {
        sessionTimersExpires: 120,
        fromUserName: userId,
        fromDisplayName: userId,
        extraHeaders: [
          `Ring-TaskId: ${taskId}`,
          `Ring-ActivityId: ${activityId}`,
          `Ring-UserId: ${userId}`,
          `Ring-Auth: ${token}`
        ],
        pcConfig: { iceServers: iceServers },
        mediaConstraints: { audio: true, video: false }
      }
      const target = `sip:${taskId}@example.com`

      const ns = this.ua.call(target, options)
      const status = this.session.status
      status.value = ''

      ns.connection.ontrack = (e) => {
        // media stream: e.streams[0]
        resolve(e.streams[0])
        console.debug('session track', e)
      }

      listenSEvs.forEach((e) => {
        ns.on(e, (data) => {
          status.value = e
          console.debug('session event', e, data)

          if (e == 'failed') {
            endFunc(data)
            reject(data)
          }
        })
      })

      this.session._session = ns
    })
  }
}

const phone = Object.create(Phone)

export default phone

const ListenPEvs = [
  //'sipEvent',
  'connecting',
  'connected',
  'disconnected'
  //'registered',
  //'unregistered',
  //'registrationFailed'
  // 'registrationExpiring'
]
const listenSEvs = [
  'peerconnection',
  'connecting',
  'sending',
  'progress',
  'accepted',
  'confirmed',
  'ended',
  'failed'
]
