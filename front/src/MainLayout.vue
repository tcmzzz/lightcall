<script setup>
import { ref } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useUserStore } from '@/stores/user'

const uStore = useUserStore()

const menu = ref()
const usermenu = ref([{ label: 'Logout', icon: 'pi pi-sign-out', command: () => uStore.logout() }])

const toggle_usermenu = (event) => {
  menu.value.toggle(event)
}
</script>

<template>
  <Toast />
  <ConfirmPopup />

  <div class="container h-full mx-auto mb-4 md:mb-6">
    <header class="flex items-center px-2 md:px-0">
      <div class="flex w-24">
        <img alt="logo" class="logo" src="@/assets/logo.png" width="100" height="100" />
      </div>
      <nav class="flex items-center ml-5">
        <RouterLink
          active-class="text-lg font-semibold"
          class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
          to="/task"
        >
          <i class="pi pi-list-check mr-2"></i>我的任务
        </RouterLink>
        <RouterLink
          v-if="uStore.userInfo?.isAdmin"
          active-class="text-lg font-semibold"
          class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
          to="/objective/table"
        >
          <i class="pi pi-bullseye mr-2"></i>外呼目标
        </RouterLink>
      </nav>

      <nav class="flex items-center ml-auto">
        <RouterLink
          v-if="uStore.userInfo?.isAdmin"
          active-class="text-lg font-semibold"
          class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
          to="/call-config/number"
        >
          <i class="pi pi-phone mr-2"></i>通话配置
        </RouterLink>
        <RouterLink
          v-if="uStore.userInfo?.isAdmin"
          active-class="text-lg font-semibold"
          class="hidden text-gray-600 hover:text-gray-800 sm:inline-block px-2 py-1"
          to="/sysconfig"
        >
          <i class="pi pi-cog mr-2"></i>系统设置
        </RouterLink>
      </nav>
      <div class="ml-10">
        <Menu id="overlay_menu" ref="menu" :model="usermenu" :popup="true" />
        <span class="mr-4 cursor-pointer" @click="toggle_usermenu">
          <i class="pi pi-user mr-2"></i>Hi {{ uStore.userInfo?.name }} ~
        </span>
      </div>
    </header>

    <div class="h-5/6 px-4 shadow border rounded-md">
      <RouterView />
    </div>
  </div>
</template>
