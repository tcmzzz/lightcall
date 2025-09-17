import './assets/main.css'
import 'primeicons/primeicons.css'

import { setLocale } from 'yup'
import { zh } from 'yup-locales'
setLocale(zh)

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import { useUserStore } from '@/stores/user'

import PrimeVue from 'primevue/config'
import Ripple from 'primevue/ripple'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'

import Aura from '@primeuix/themes/aura'

import App from './App.vue'
import router from './router'
const app = createApp(App)

app.use(PrimeVue, {
  ripple: true,
  theme: {
    preset: Aura
  }
})

app.directive('ripple', Ripple)
app.use(ConfirmationService)
app.use(ToastService)
app.use(createPinia())
app.use(router)

const uStore = useUserStore()

app.mount('#app')

router.beforeEach(async (to, from) => {
  if (!uStore.isValid && to.name !== 'login') {
    // 将用户重定向到登录页面
    return '/login'
  }
  if (to.name == 'login' && uStore.isValid) {
    return from
  }
})
