<script setup>
const items = [
  { route: '/sysconfig/config', label: '系统配置', icon: 'pi pi-cog' },
  { route: '/sysconfig/user', label: '用户管理', icon: 'pi pi-users' }
]
</script>

<template>
  <SubMenu :menuitems="items" />
</template>
