<script setup>
import Tag from 'primevue/tag'
import { useToast } from 'primevue/usetoast'
import { ref, computed } from 'vue'
import PhoneCard from '@/components/task/PhoneCard.vue'
import ObjectiveInfo from '@/components/ObjectiveInfo.vue'
import Textarea from 'primevue/textarea'
import Button from 'primevue/button'
import { pb } from '@/pocketbase/index'
import { useUserStore } from '@/stores/user'
import { renderMarkdown } from '@/util/markdown'

const toast = useToast()
const loading = ref(false)
const showPhoneCard = ref(false)
const showObjectiveInfo = ref(false)
const isEditingDescription = ref(false)
const editableDescription = ref('')
const userStore = useUserStore()

const props = defineProps({
  taskData: {
    type: Object,
    required: true
  },
  users: {
    type: Array,
    default: () => []
  },
  objective: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'need-refresh'])

// 检查当前用户是否为管理员
const isAdmin = computed(() => {
  return userStore.userInfo?.isAdmin || false
})

// 渲染Markdown格式的背景信息
const renderedBackground = computed(() => {
  const background = props.objective?.info?.background || ''
  return renderMarkdown(background)
})

// 渲染Markdown格式的任务描述
const renderedTaskDescription = computed(() => {
  const description = props.taskData?.desc || ''
  return renderMarkdown(description)
})

// 统一更新任务字段
async function updateTask(field, value) {
  try {
    loading.value = true
    const data = await pb.collection('task').update(props.taskData.id, {
      [field]: value
    })

    emit('need-refresh')

    const msg = data.message ? '.' + data.message : ''
    toast.add({
      severity: 'success',
      summary: '更新成功',
      detail: field === 'own' ? '负责人已变更' + msg : '状态已更新' + msg,
      life: 3000
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '更新失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}

// 更新负责人
const updateOwner = (newId) => updateTask('own', newId)
// 切换任务状态
const toggleTask = (flag) => updateTask('open', flag)

// 开始编辑任务描述
function startEditDescription() {
  editableDescription.value = props.taskData?.desc || ''
  isEditingDescription.value = true
}

// 取消编辑
function cancelEditDescription() {
  isEditingDescription.value = false
  editableDescription.value = ''
}

// 保存任务描述
async function updateDescription() {
  if (!props.taskData) return
  try {
    loading.value = true
    await pb.collection('task').update(props.taskData.id, {
      desc: editableDescription.value
    })

    emit('need-refresh')

    toast.add({
      severity: 'success',
      summary: '更新成功',
      detail: '任务描述已更新',
      life: 3000
    })

    isEditingDescription.value = false
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '更新失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-4 p-4 h-full overflow-y-auto">
    <!-- 第一行：目标名称和公司名称 -->
    <div class="flex flex-col md:flex-row gap-4">
      <div class="flex-1">
        <label class="text-sm text-gray-500">目标名称</label>
        <p class="font-medium">
          {{ objective?.title }}

          <Tag
            :value="taskData?.open ? '进行中' : '已关闭'"
            :severity="taskData?.open ? 'success' : 'danger'"
          />
        </p>
      </div>
      <div class="flex-1">
        <label class="text-sm text-gray-500">公司名称</label>
        <p class="font-medium">{{ objective?.info?.company }}</p>
      </div>
    </div>

    <!-- 第二行：负责人、状态和操作 -->
    <div class="flex flex-col md:flex-row items-center">
      <div class="flex-1 flex flex-col">
        <label class="text-sm text-gray-500">负责人</label>
        <Select
          v-if="isAdmin && users?.length > 0"
          class="w-48"
          :model-value="taskData.own"
          :options="users"
          option-label="name"
          option-value="id"
          placeholder="选择负责人"
          @update:model-value="(val) => updateOwner(val)"
        />
        <span v-else>{{ taskData.expand?.own?.name || taskData.own }}</span>
      </div>

      <div class="flex-1 items-center">
        <ConfirmBtn
          v-tooltip="taskData.ext_id && taskData.id !== taskData.ext_id ? '需待源系统处理' : ''"
          :label="taskData.open ? '关闭任务' : '开启任务'"
          severity="info"
          text
          :confirm-msg="`确定要${taskData.open ? '关闭' : '开启'}此任务吗？`"
          @confirm-act="() => toggleTask(!taskData.open)"
        />
        <Button v-if="taskData?.open" @click="showPhoneCard = true">呼叫</Button>
      </div>
    </div>

    <!-- 第三行：客户信息和联系方式 -->
    <div class="flex flex-col md:flex-row gap-4">
      <div class="flex-1">
        <label class="text-sm text-gray-500">客户信息</label>
        <p class="font-medium">{{ taskData.contact }}</p>
      </div>
      <div class="flex-1">
        <label class="text-sm text-gray-500">联系方式</label>
        <p class="font-medium">{{ taskData.callee }}</p>
      </div>
    </div>

    <!-- 第四行：背景信息 -->
    <div>
      <label class="text-sm text-gray-500">背景信息</label>
      <div class="font-medium text-sm bg-gray-50 p-2 rounded">
        <!-- eslint-disable vue/no-v-html -->
        <div class="max-h-32 overflow-y-auto">
          <div class="prose prose-sm max-w-none" v-html="renderedBackground"></div>
        </div>
        <!-- eslint-enable vue/no-v-html -->
        <Button
          label="查看详情"
          text
          size="small"
          class="p-0 mt-1"
          @click="showObjectiveInfo = true"
        />
      </div>
    </div>

    <!-- 第五行：任务描述 -->
    <div class="flex-1 flex flex-col min-h-0">
      <div class="flex items-center justify-between mb-1">
        <label class="text-sm text-gray-500">任务描述</label>
        <Button
          v-if="isAdmin && !isEditingDescription"
          icon="pi pi-pencil"
          class="p-button-rounded p-button-text"
          @click="startEditDescription"
        />
      </div>

      <template v-if="isEditingDescription">
        <Textarea
          v-model="editableDescription"
          class="w-full text-sm h-60"
          placeholder="支持Markdown格式：\n**粗体**、*斜体*、[链接](url)、- 列表"
        />
        <span class="text-sm text-gray-500"
          >(支持Markdown格式：\n**粗体**、*斜体*、[链接](url)、- 列表)</span
        >
        <div class="flex justify-end gap-2 mt-2">
          <Button label="取消" severity="secondary" size="small" @click="cancelEditDescription" />
          <Button label="保存" size="small" :loading="loading" @click="updateDescription" />
        </div>
      </template>

      <template v-else>
        <!-- eslint-disable vue/no-v-html -->
        <div class="font-medium bg-gray-50 p-2 rounded h-full overflow-y-auto">
          <div class="prose prose-sm max-w-none" v-html="renderedTaskDescription"></div>
        </div>
        <!-- eslint-enable vue/no-v-html -->
      </template>
    </div>
  </div>

  <!-- 电话卡片组件 -->
  <PhoneCard v-model:visible="showPhoneCard" :task-id="taskData.id" />

  <!-- 目标详情组件 -->
  <ObjectiveInfo
    v-model:visible="showObjectiveInfo"
    :object="objective"
    @updated="emit('need-refresh')"
  />
</template>
