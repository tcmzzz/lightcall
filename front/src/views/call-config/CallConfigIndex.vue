<script setup>
const items = [
  { route: '/call-config/number', label: '可用号码', icon: 'pi pi-phone' },
  { route: '/call-config/gateway', label: '呼出网关', icon: 'pi pi-sitemap' }
]
</script>

<template>
  <SubMenu :menuitems="items" />
</template>
