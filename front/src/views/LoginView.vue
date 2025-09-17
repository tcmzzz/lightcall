<script setup>
import { useUserStore } from '@/stores/user'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useToast } from 'primevue/usetoast'
import Password from 'primevue/password'
import Toast from 'primevue/toast'

const uStore = useUserStore()
const router = useRouter()
const toast = useToast()

const username = ref('')
const passwd = ref('')
const isLoading = ref(false)

async function login() {
  if (!username.value || !passwd.value) return

  try {
    isLoading.value = true
    await uStore.login(username.value, passwd.value)
    await router.push('/')
  } catch (e) {
    toast.add({
      severity: 'error',
      summary: '登录失败',
      life: 3000
    })
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-slate-100 to-blue-50 flex items-center p-4">
    <Toast position="top-center" />
    <div class="mx-auto w-full max-w-md bg-white rounded-xl shadow-lg overflow-hidden">
      <div class="p-8 space-y-6">
        <div class="text-center">
          <div class="inline-block bg-blue-100 p-4 rounded-2xl mb-6">
            <i class="pi pi-user text-4xl text-blue-600" />
          </div>
          <h1 class="text-3xl font-bold text-gray-800 mb-2">欢迎回来</h1>
          <p class="text-gray-500">请输入您的登录凭证</p>
        </div>

        <form class="space-y-4" @submit.prevent="login">
          <div class="relative w-full">
            <InputText
              id="username"
              v-model="username"
              class="w-full px-4 py-3 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent peer transition-colors duration-200"
              :class="{ '!border-blue-500': username }"
            />
            <label
              for="username"
              class="absolute left-4 transition-all duration-200 pointer-events-none origin-top-left"
              :class="
                username
                  ? 'top-1.5 text-xs text-blue-600 scale-90'
                  : 'top-3.5 text-gray-400 scale-100'
              "
            >
              用户名
            </label>
          </div>

          <div class="relative w-full">
            <label
              for="password"
              class="absolute left-4 transition-all duration-200 pointer-events-none origin-top-left"
              :class="
                passwd
                  ? 'top-1.5 text-xs text-blue-600 scale-90'
                  : 'top-3.5 text-gray-400 scale-100'
              "
            >
              密码
            </label>
            <Password
              id="password"
              v-model="passwd"
              toggle-mask
              input-class="w-full px-4 py-3 pr-12 border border-gray-200 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors duration-200"
              :class="{ '!border-blue-500': passwd }"
              panel-class="p-4 bg-white rounded-lg shadow-lg border border-gray-200 w-full"
              :feedback="false"
            />
          </div>

          <Button
            type="submit"
            label="登录系统"
            class="w-full !py-3 !bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 !text-white !font-semibold !rounded-lg transition-all"
            :loading="isLoading"
            :disabled="isLoading"
          >
            <template #icon>
              <i class="pi pi-arrow-right-to-bracket mr-2" />
            </template>
          </Button>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
input:-webkit-autofill,
input:-webkit-autofill:hover,
input:-webkit-autofill:focus {
  -webkit-box-shadow: 0 0 0 100px white inset;
  -webkit-text-fill-color: #1e293b;
}

.p-inputtext {
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease !important;
}

.p-password input {
  padding-right: 3.5rem;
}

.p-password-panel .p-password-meter {
  display: none;
}

.p-password {
  width: 100%;
}

.p-password-toggle {
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
}

label {
  transform-origin: left center;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
