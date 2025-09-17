<script setup>
import { ref, watch } from 'vue'
import phone from '@/util/phone'
import { useToast } from 'primevue/usetoast'
import { pb } from '@/pocketbase'

const props = defineProps({
  taskId: {
    type: String,
    required: true
  },
  visible: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['update:visible'])

const toast = useToast()
const audioRef = ref(null)
const callStatus = ref('准备中...')
const localVisible = ref(props.visible)
const taskInfo = ref(null)
const activityId = ref(null)

// 步骤状态
const steps = ref([
  {
    label: '创建活动',
    status: 'wait',
    error: '',
    fn: async () => {
      try {
        updateStepStatus(0, 'running')
        await new Promise((resolve) => setTimeout(resolve, 500))
        const obj = await pb.send('/api/custom/call/new/' + props.taskId, { query: {} })
        activityId.value = obj.id
        updateStepStatus(0, 'success')
        return true
      } catch (error) {
        updateStepStatus(0, 'error', error.message || '创建活动失败')
        return false
      }
    }
  },
  {
    label: '黑名单查询',
    status: 'wait',
    error: '',
    fn: async () => {
      return invokePreCall(1, '/api/custom/call/precall/blacklist')
    }
  },
  {
    label: '闪信发送',
    status: 'wait',
    error: '',
    fn: async () => {
      return invokePreCall(2, '/api/custom/call/precall/flashcard')
    }
  },
  {
    label: '启动呼叫',
    status: 'wait',
    error: '',
    fn: async () => {
      try {
        updateStepStatus(3, 'running')
        callStatus.value = '正在初始化...'
        await phone.init()
        await phone.start()

        await new Promise((resolve) => setTimeout(resolve, 500))

        callStatus.value = '正在呼叫...'
        const stream = await phone.call(props.taskId, activityId.value, function (data) {
          console.log('in end func!!!!!')
          callStatus.value = '通话结束'
          toast.add({
            severity: 'warn',
            summary: '被叫挂断',
            detail: '呼叫过程中出错(' + data.cause + ')',
            life: 3000
          })
        })

        if (stream && audioRef.value) {
          audioRef.value.srcObject = stream
          callStatus.value = '通话中...'
          updateStepStatus(3, 'success')
        }
        return true
      } catch (error) {
        console.log('in catch!!!!!')
        callStatus.value = '呼叫失败'
        updateStepStatus(3, 'error', error.message || '启动呼叫失败')
        toast.add({
          severity: 'error',
          summary: '呼叫失败',
          detail: error.message || '呼叫过程中出错',
          life: 3000
        })
        return false
      }
    }
  }
])

async function invokePreCall(index, url) {
  try {
    updateStepStatus(index, 'running')

    await new Promise((resolve) => setTimeout(resolve, 500))

    const obj = await pb.send(url + '/' + activityId.value, {
      query: {}
    })
    if (obj.pass === true) {
      updateStepStatus(index, 'success', obj.msg)
      return true
    } else {
      updateStepStatus(index, 'error', obj.msg)
      return false
    }
  } catch (error) {
    updateStepStatus(index, 'error', error.message || '接口调用失败')
    return false
  }
}
// 重置步骤状态
function resetSteps() {
  steps.value.forEach((step) => {
    step.status = 'wait'
    step.error = ''
  })
}

// 更新步骤状态
function updateStepStatus(index, status, error = '') {
  steps.value[index].status = status
  steps.value[index].error = error
}

// 执行指定步骤
async function invokeStep(index) {
  if (index >= steps.value.length) return
  const step = steps.value[index]
  const success = await step.fn()
  if (success && index < steps.value.length - 1) {
    await invokeStep(index + 1)
  }
}

// 顺序执行所有步骤
async function executeSteps() {
  resetSteps()
  activityId.value = null
  await invokeStep(0)
}

// 获取任务信息
async function fetchTaskInfo() {
  try {
    taskInfo.value = await pb.collection('task').getOne(props.taskId, {
      expand: 'objective'
    })
  } catch (error) {
    console.error('获取任务信息失败:', error)
  }
}

// 监听visible变化
watch(
  () => props.visible,
  async (val) => {
    localVisible.value = val
    if (val) {
      await fetchTaskInfo()
      executeSteps()
    } else {
      endCall()
    }
  }
)

// 结束呼叫
function endCall() {
  phone.hangup()
  callStatus.value = '已挂断'
  if (audioRef.value) {
    audioRef.value.srcObject = null
  }
}

// 挂断电话
function hangup() {
  endCall()
  emit('update:visible', false)
}
</script>

<template>
  <Dialog
    v-model:visible="localVisible"
    :modal="false"
    header="通话中"
    :style="{ width: '500px' }"
    :closable="false"
    position="top"
  >
    <div class="flex flex-col gap-4">
      <!-- 步骤状态 -->
      <div class="flex justify-between mb-4">
        <div v-for="(step, index) in steps" :key="index" class="flex flex-col items-center">
          <div class="text-2xl">
            <i v-if="step.status === 'success'" class="pi pi-check-circle text-green-500" />
            <i v-else-if="step.status === 'error'" class="pi pi-times-circle text-red-500" />
            <i v-else-if="step.status === 'running'" class="pi pi-spinner animate-spin" />
            <i v-else class="pi pi-circle" />
          </div>
          <div class="text-sm font-medium">{{ step.label }}</div>
          <small v-if="step.status === 'error'" class="text-red-500">
            {{ step.error }}
            <span
              v-if="index < steps.length - 1"
              size="small"
              class="text-gray-500 cursor-pointer"
              @click="invokeStep(index + 1)"
            >
              (忽略)
            </span>
          </small>
          <small v-else class="text-green-500">
            {{ step.error }}
          </small>
        </div>
      </div>

      <!-- 任务信息 -->
      <div class="p-3 border-round surface-100">
        <div v-if="taskInfo" class="space-y-2">
          <div class="text-lg font-semibold">{{ taskInfo.contact }}</div>
          <div class="text-gray-600">{{ taskInfo.callee }}</div>
          <div v-if="taskInfo.expand?.objective" class="text-sm text-gray-500">
            {{ taskInfo.expand.objective.title }}
          </div>
        </div>
      </div>

      <!-- 通话状态 -->
      <div class="p-3 border-round surface-100">
        <div class="flex items-center gap-2">
          <i class="pi pi-phone" />
          <span>{{ callStatus }}</span>
        </div>
      </div>

      <!-- 音频控制 -->
      <div class="p-3 border-round surface-100">
        <audio ref="audioRef" controls autoplay class="w-full" />
      </div>

      <!-- 操作按钮 -->
      <div class="flex justify-center">
        <Button label="挂断" severity="danger" icon="pi pi-phone" @click="hangup" />
      </div>
    </div>
  </Dialog>
</template>

<style scoped>
.animate-spin {
  animation: spin 1s linear infinite;
}
</style>
