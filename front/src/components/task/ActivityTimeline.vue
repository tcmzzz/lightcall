<script setup>
import { useUserStore } from '@/stores/user'
import { formatAbsoluteTime, formatRelativeTime } from '@/util/time'
import { useToast } from 'primevue/usetoast'
import { pb } from '@/pocketbase/index'
import { ref, onMounted, nextTick, computed, watch } from 'vue'

const uStore = useUserStore()
const toast = useToast()
const props = defineProps({
  activities: { type: Object, required: true },
  taskId: { type: String, required: true }
})

const comment = ref('')
const emit = defineEmits(['need-refresh'])
const loading = ref(false)
const activityListRef = ref(null)

// 按时间排序活动，最早的在前，最新的在后
const sortedActivities = computed(() => {
  if (!props.activities || !Array.isArray(props.activities)) return []
  return [...props.activities].sort((a, b) => new Date(a.created) - new Date(b.created))
})

// 滚动到底部显示最新活动
const scrollToBottom = async () => {
  await nextTick()
  if (activityListRef.value) {
    activityListRef.value.scrollTop = activityListRef.value.scrollHeight
  }
}

// 组件挂载后滚动到底部
onMounted(() => {
  scrollToBottom()
})

// 活动更新时也滚动到底部
watch(
  () => props.activities,
  () => {
    scrollToBottom()
  },
  { deep: true }
)

const getRecordUrl = (activity) => {
  return pb.files.getURL(activity, activity.record, { token: pb.files.getToken() })
}

watch(
  () => props.taskId,
  async (newId, oldId) => {
    if (oldId) {
      await pb.collection('task').unsubscribe(oldId)
    }
    if (newId) {
      await pb.collection('task').subscribe(newId, function () {
        emit('need-refresh')
      })
    }
  },
  { immediate: true }
)

async function addComment() {
  const c = comment.value.trim()
  if (c === '') return

  try {
    // 1. 创建新的 activity 记录
    loading.value = true
    const newActivity = await pb.collection('activity').create({
      user: uStore.userInfo.id,
      comment: c,
      isCall: false,
      task: props.taskId
    })

    // 2. 将新 activity 关联到 task
    await pb.collection('task').update(props.taskId, {
      'activity+': newActivity.id
    })

    emit('need-refresh')
    comment.value = ''
    toast.add({
      severity: 'success',
      summary: '回复成功',
      detail: '新活动记录已添加',
      life: 3000
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '提交失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="h-full flex flex-col overflow-hidden gap-5">
    <label class="text-gray-500 shrink-0">活动列表</label>

    <!-- 活动列表区域 -->
    <div ref="activityListRef" class="flex-1 min-h-0 overflow-y-auto bg-gray-50 space-y-4 p-2">
      <Card
        v-for="(activity, index) in sortedActivities"
        :key="activity.id"
        class="shadow-sm hover:shadow-md transition-shadow max-w-[400px]"
        :class="index % 2 === 0 ? 'mr-auto' : 'ml-auto'"
      >
        <template #header>
          <div class="flex items-center justify-between p-2">
            <span class="text-sm font-medium">
              {{ activity.expand?.user?.name || '系统记录' }}
            </span>
            <div class="flex items-center gap-2">
              <span class="text-xs text-gray-500">
                {{ formatAbsoluteTime(activity.created) }}
                <span class="ml-1">({{ formatRelativeTime(activity.created) }})</span>
              </span>
              <Tag v-if="activity.isCall" value="通话" severity="info" class="text-xs" />
            </div>
          </div>
        </template>

        <template #content>
          <div class="p-2">
            <p class="text-sm">{{ activity.comment }}</p>

            <div v-if="activity.record" class="mt-3">
              <audio controls class="w-full">
                <source :src="getRecordUrl(activity)" type="audio/mpeg" />
                您的浏览器不支持音频播放
              </audio>
            </div>
          </div>
        </template>
      </Card>

      <div v-if="sortedActivities.length === 0" class="text-center py-8 text-gray-400">
        暂无活动记录
      </div>
    </div>

    <!-- 评论输入区域 -->
    <div class="h-[160px] shrink-0 flex flex-col gap-1">
      <div class="flex">
        <Textarea
          v-model="comment"
          rows="4"
          placeholder="输入回复内容..."
          auto-resize
          class="w-full h-full text-sm"
        />
      </div>

      <div class="flex">
        <Button
          label="提交回复"
          size="small"
          class="ml-auto"
          :disabled="!comment.trim()"
          @click="addComment"
        />
      </div>
    </div>
  </div>
</template>
