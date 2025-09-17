import { defineStore } from 'pinia'
import { useRouter } from 'vue-router'
import phone from '@/util/phone'

export const useUserStore = defineStore('user', () => {
  const router = useRouter()

  const userInfo = ref(pb.authStore.model)
  const isValid = ref(pb.authStore.isValid)
  function fresh() {
    userInfo.value = pb.authStore.model
    isValid.value = pb.authStore.isValid
  }
  async function login(username, passwd) {
    try {
      await pb.collection('users').authWithPassword(username, passwd)
      fresh()
      router.push('/')
    } catch (e) {
      throw new Error('登陆失败')
    }
  }

  function logout() {
    phone.stop()
    pb.authStore.clear()
    fresh()
    router.push('/login')
  }

  async function checkAuth() {
    if (pb.authStore?.isValid) {
      try {
        await pb.collection('users').authRefresh()
      } catch (e) {
        console.error('check auth fail', e)
        logout()
      }
    }
  }

  // check auth when load store(when fresh page)
  checkAuth()

  return { userInfo, isValid, login, logout }
})
