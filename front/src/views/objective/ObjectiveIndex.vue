<script setup>
import { provide, ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const objectiveTitle = ref('')
const showObjectiveDetail = ref(false)

const showBreadcrumbs = computed(() => {
  return route.path.includes('/detail/')
})

// 提供目标标题给子组件
provide('objectiveTitle', objectiveTitle)
provide('showObjectiveDetail', showObjectiveDetail)
</script>

<template>
  <div class="flex flex-col h-full">
    <div v-if="showBreadcrumbs" class="flex items-center justify-between px-4 py-2 mt-5">
      <nav class="flex items-center gap-2 text-sm">
        <RouterLink to="/objective/table" class="text-primary-600 hover:text-primary-800">
          目标列表
        </RouterLink>
        <i class="pi pi-chevron-right text-xs text-gray-400" />
        <span v-if="objectiveTitle" class="text-gray-600">{{ objectiveTitle }}</span>
      </nav>
      <Button
        v-if="objectiveTitle"
        text
        severity="secondary"
        label="查看详情"
        icon="pi pi-eye"
        size="small"
        class="mr-auto"
        @click="showObjectiveDetail = true"
      />
    </div>

    <RouterView />
  </div>
</template>
