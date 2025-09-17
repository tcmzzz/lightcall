<script setup>
import { ref, computed, watch } from 'vue'
import ActivityTimeline from '@/components/task/ActivityTimeline.vue'
import InfoPanel from '@/components/task/InfoPanel.vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  tasks: {
    type: Array,
    required: true
  },
  initialIndex: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['update:visible', 'need-refresh'])

const dialog = ref()
const currentIndex = ref(props.initialIndex)
const localVisible = ref(props.visible)

watch(
  () => props.visible,
  (val) => {
    localVisible.value = val
    if (localVisible.value) {
      if (dialog.value.maximized) return
      dialog.value.maximize()
    }
  }
)

const currentTask = computed(() => props.tasks[currentIndex.value])

const sortedActivities = computed(() => {
  return (
    currentTask.value?.expand?.activity
      ?.slice()
      .sort((a, b) => new Date(b.created) - new Date(a.created)) || []
  )
})

const switchTask = (step) => {
  const newIndex = currentIndex.value + step
  if (newIndex < 0) {
    currentIndex.value = props.tasks.length - 1
  } else if (newIndex >= props.tasks.length) {
    currentIndex.value = 0
  } else {
    currentIndex.value = newIndex
  }
}
</script>

<template>
  <Dialog
    ref="dialog"
    v-model:visible="localVisible"
    :modal="true"
    :closable="true"
    :closeOnEscape="false"
    :maximizable="true"
    class="z-999"
    @hide="$emit('update:visible', false)"
  >
    <template #header>
      <div class="flex items-center justify-between p-4 border-b">
        <span class="text-lg text-gray-600">
          当前任务：{{ currentIndex + 1 }} / {{ tasks.length }}
        </span>
        <!--
        <Button text severity="secondary" label="关闭" @click="$emit('update:visible', false)" />
    -->
      </div>
    </template>

    <div class="flex flex-col h-full">
      <!-- 顶部操作栏 -->

      <main class="flex-1 flex overflow-hidden">
        <!-- 左侧切换按钮 -->
        <button
          class="w-16 flex items-center justify-center hover:bg-gray-50 cursor-pointer"
          @click="switchTask(-1)"
        >
          <i class="pi pi-chevron-left text-3xl text-primary-500" />
        </button>

        <!-- 主要内容 -->
        <div class="flex-1 flex flex-col md:flex-row overflow-hidden">
          <!-- 任务信息 -->
          <div class="flex-shrink-1 min-w-[500px] w-full md:w-1/2 flex flex-col p-4">
            <!-- flex-grow填充剩余空间 -->
            <div class="flex-grow overflow-y-auto">
              <InfoPanel
                :task-data="currentTask"
                :users="[]"
                :objective="currentTask.objective"
                @need-refresh="$emit('need-refresh')"
              />
            </div>

            <!-- flex-shrink-0防止压缩 -->
            <div class="flex-shrink-0 mt-4 min-h-[100px]"></div>
          </div>

          <!-- 活动时间线 -->
          <div class="flex-shrink-2 w-full md:w-1/2 overflow-y-auto p-4">
            <ActivityTimeline
              :activities="sortedActivities"
              :task-id="currentTask?.id"
              @need-refresh="$emit('need-refresh')"
            />
          </div>
        </div>

        <!-- 右侧切换按钮 -->
        <button
          class="w-16 flex items-center justify-center hover:bg-gray-50 cursor-pointer"
          @click="switchTask(1)"
        >
          <i class="pi pi-chevron-right text-3xl text-primary-500" />
        </button>
      </main>
    </div>
  </Dialog>
</template>
