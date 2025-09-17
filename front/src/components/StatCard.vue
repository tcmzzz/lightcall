<script setup>
defineProps({
  title: { type: String, default: '' },
  value: { type: [String, Number], default: 0 },
  icon: { type: Object, default: () => {} },
  suffix: { type: String, default: '' }
})
</script>

<template>
  <div
    class="p-4 border-round-lg surface-card shadow-2 hover:shadow-4 transition-all transition-duration-300 cursor-pointer bg-gradient-to-r from-white to-gray-50 hover:to-gray-100"
  >
    <div class="flex align-items-center gap-4">
      <!-- 图标容器 -->

      <div class="p-3 border-round-lg flex items-center justify-center">
        <component :is="icon" class="w-10 h-10" />
      </div>

      <!-- 内容区域 -->
      <div class="flex flex-column">
        <span class="text-gray-500 font-medium tracking-wide">{{ title }}</span>
        <!-- 数字专用区域 -->
        <div class="mt-2 p-3">
          <div class="flex flex-col items-center gap-1">
            <span class="text-4xl font-bold leading-none">{{ value }}</span>
            <span v-if="suffix" class="text-sm font-medium">{{ suffix }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
