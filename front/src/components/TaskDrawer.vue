<script setup>
import { ref, watch, computed } from 'vue'
import InfoPanel from '@/components/task/InfoPanel.vue'
import { useToast } from 'primevue/usetoast'
import ActivityTimeline from '@/components/task/ActivityTimeline.vue'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  taskData: {
    type: Object,
    default: null
  },
  objective: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'need-refresh'])

const toast = useToast()
const users = ref([])
const loading = ref(false)
const localVisible = ref(props.visible)

watch(
  () => props.visible,
  async (val) => {
    localVisible.value = val
    if (val) {
      await loadUsers()
    }
  }
)
// 加载用户列表
async function loadUsers() {
  try {
    loading.value = true
    users.value = await pb.collection('users').getFullList({ sort: 'name' })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '加载用户失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}

const sortedActivities = computed(() => {
  return (
    props.taskData?.expand?.activity
      ?.slice()
      .sort((a, b) => new Date(b.created) - new Date(a.created)) || []
  )
})
</script>

<template>
  <Drawer
    v-model:visible="localVisible"
    position="left"
    :style="{ width: 'min(90vw, 1200px)' }"
    :modal="true"
    :closeOnEscape="false"
    class="task-drawer-with-comment"
    @hide="$emit('update:visible', false)"
  >
    <template #container>
      <div v-if="taskData" class="mr-5 flex flex-col h-full" style="overflow: hidden">
        <!-- 两列布局：InfoPanel在左，ActivityTimeline在右 -->
        <div class="flex flex-1 gap-4 h-full flex-col lg:flex-row" style="overflow: hidden">
          <!-- 左列：InfoPanel -->
          <div class="flex-1 overflow-auto">
            <InfoPanel
              :task-data="taskData"
              :users="users"
              :objective="objective"
              @need-refresh="emit('need-refresh')"
            />
          </div>

          <!-- 右列：ActivityTimeline -->
          <div class="flex-1 overflow-hidden">
            <ActivityTimeline
              :activities="sortedActivities"
              :task-id="props.taskData.id"
              @need-refresh="emit('need-refresh')"
            />
          </div>
        </div>
      </div>
    </template>
  </Drawer>
</template>
