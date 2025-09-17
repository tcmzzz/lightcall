<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { pb } from '@/pocketbase/index'
import { useToast } from 'primevue/usetoast'
import { useUserStore } from '@/stores/user'

import TaskDrawer from '@/components/TaskDrawer.vue'
import FocusModeDrawer from '@/components/task/FocusModeDrawer.vue'
import ConfirmBtn from '@/components/ConfirmBtn.vue'
import { formatRelativeTime } from '@/util/time'
import { renderMarkdownInline } from '@/util/markdown'

const toast = useToast()
const userStore = useUserStore()
const allTasks = ref([]) // 存储所有任务数据
const tasks = ref([])
const drawerVisible = ref(false)
const selectedTask = ref(null)
const selectedTasks = ref([])

const focusDrawerVisible = ref(false)
const focusTasks = ref([])
const initialFocusIndex = ref(0)

const directDialNumber = ref('') // 直接拨号输入框的值

// 渲染Markdown格式的任务描述（行内模式）
const renderedTaskDescriptionInline = (desc) => {
  return renderMarkdownInline(desc || '')
}

// 新增：状态筛选变量
const statusFilter = ref('open') // 默认只显示开启的任务
const statusOptions = [
  { label: '开启', value: 'open' },
  { label: '关闭', value: 'closed' },
  { label: '全部', value: 'all' }
]

// 客户端过滤任务数据
const filteredTasks = computed(() => {
  if (statusFilter.value === 'all') return allTasks.value
  return allTasks.value.filter((task) => (statusFilter.value === 'open' ? task.open : !task.open))
})

// 监听状态筛选变化
watch(statusFilter, () => {
  // 不再需要调用 loadData()，使用客户端过滤
})

// 处理直接拨号
const handleDirectDial = async () => {
  if (!directDialNumber.value) {
    toast.add({
      severity: 'warn',
      summary: '请输入号码',
      detail: '请先输入要拨打的电话号码',
      life: 3000
    })
    return
  }

  try {
    // 调用新的服务端接口创建任务
    const response = await pb.send('/api/custom/call/direct', {
      method: 'POST',
      body: {
        number: directDialNumber.value
      }
    })

    // 获取创建的任务
    const task = await pb.collection('task').getOne(response.taskId, {
      expand: 'own, activity.user, objective_via_tasks'
    })

    // 进入专注模式
    focusTasks.value = [task]
    await loadData()
    initialFocusIndex.value = 0
    focusDrawerVisible.value = true

    // 清空输入框
    directDialNumber.value = ''
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '拨号失败',
      detail: error.message,
      life: 3000
    })
  }
}

const enterFocusMode = () => {
  if (selectedTasks.value.length === 0) {
    toast.add({
      severity: 'warn',
      summary: '请选择任务',
      detail: '请至少选择一个任务进入专注模式',
      life: 3000
    })
    return
  }
  focusTasks.value = selectedTasks.value
  initialFocusIndex.value = 0
  focusDrawerVisible.value = true
}

const getLatestActivity = (task) => {
  const activities = task.expand?.activity || []
  return activities[activities.length - 1] || null
}

async function loadData() {
  try {
    // 获取当前用户的所有任务（不进行状态过滤）
    const filter = `own="${userStore.userInfo.id}"`

    const list = await pb.collection('task').getFullList({
      filter: filter,
      expand: 'own, activity.user, objective_via_tasks',
      sort: '-created'
    })

    const newTasks = list.map((task) => ({
      ...task,
      objective: task.expand.objective_via_tasks?.[0],
      latestActivity: getLatestActivity(task)?.created || '1970-01-01'
    }))

    // 更新所有任务数据
    allTasks.value = newTasks
    // 更新显示的任务（通过 computed filteredTasks 自动过滤）

    // 更新 focusTasks 中存在的任务数据
    if (focusTasks.value.length > 0) {
      focusTasks.value = focusTasks.value.map((oldTask) => {
        const updatedTask = allTasks.value.find((t) => t.id === oldTask.id)
        return updatedTask || oldTask
      })
    }

    if (selectedTask.value) {
      selectedTask.value = allTasks.value.find((t) => t.id === selectedTask.value.id)
    }
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '加载失败',
      detail: error.message,
      life: 3000
    })
  }
}

onMounted(() => {
  loadData()
})

// 切换任务状态
const toggleTaskStatus = (task) => {
  const newStatus = !task.open
  pb.collection('task')
    .update(task.id, { open: newStatus })
    .then((data) => {
      const msg = data.message ? '.' + data.message : ''
      toast.add({
        severity: 'success',
        summary: '操作成功',
        detail: `任务已${newStatus ? '开启' : '关闭'}${msg}`,
        life: 3000
      })
      loadData()
    })
    .catch((error) => {
      toast.add({
        severity: 'error',
        summary: '操作失败',
        detail: error.message,
        life: 3000
      })
    })
}

// 查看任务详情
const viewTaskDetail = (task) => {
  selectedTask.value = task
  drawerVisible.value = true
}

const btnMenu = [
  {
    label: '所有任务',
    command: () => {
      const openTasks = allTasks.value.filter((t) => t.open)
      if (openTasks.length === 0) {
        toast.add({
          severity: 'warn',
          summary: '没有开启的任务',
          detail: '当前没有开启的任务',
          life: 3000
        })
        return
      }
      focusTasks.value = openTasks
      initialFocusIndex.value = 0
      focusDrawerVisible.value = true
    }
  },
  {
    label: '最近24小时无活动的任务',
    command: () => {
      const oneDayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000)
      const inactiveTasks = allTasks.value.filter((t) => {
        const lastActivity = getLatestActivity(t)
        return lastActivity ? new Date(lastActivity.created) < oneDayAgo : true
      })

      if (inactiveTasks.length === 0) {
        toast.add({
          severity: 'warn',
          summary: '没有符合条件的任务',
          detail: '所有任务最近24小时都有活动',
          life: 3000
        })
        return
      }

      focusTasks.value = inactiveTasks
      initialFocusIndex.value = 0
      focusDrawerVisible.value = true
    }
  }
]
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="mt-5 mb-4 flex items-center gap-3">
      <!-- 新增：状态筛选下拉菜单 -->
      <div class="flex items-center">
        <label class="mr-2 text-gray-500">状态:</label>
        <Dropdown
          v-model="statusFilter"
          :options="statusOptions"
          option-label="label"
          option-value="value"
          class="w-32"
        />
      </div>

      <div class="flex items-center gap-2 p-inputgroup">
        <InputText v-model="directDialNumber" placeholder="输入电话号码" class="w-64" />
        <Button
          label="直接拨打"
          severity="info"
          :disabled="!directDialNumber"
          @click="handleDirectDial"
        />
      </div>

      <SplitButton
        class="ml-auto"
        label="专注模式"
        severity="info"
        :model="btnMenu"
        @click="enterFocusMode"
      />
    </div>

    <DataTable
      v-model:selection="selectedTasks"
      :value="filteredTasks"
      paginator
      :rows="7"
      scrollable
      scroll-height="flex"
      selection-mode="multiple"
      data-key="id"
      :sort-field="'latestActivity'"
      :sort-order="-1"
    >
      <Column selection-mode="multiple" style="width: 30px" />
      <Column field="open" header="状态" style="width: 80px">
        <template #body="slotProps">
          <Tag v-if="slotProps.data.open === true" severity="success"> 开启 </Tag>
          <Tag v-else severity="danger"> 关闭 </Tag>
        </template>
      </Column>
      <Column field="expand.own.name" header="负责人" style="width: 100px" sortable />

      <Column field="objective.title" header="关联目标" sortable style="width: 200px">
        <template #body="slotProps">
          <div v-if="slotProps.data.objective" class="flex flex-col">
            <span class="font-medium">{{ slotProps.data.objective.title }}</span>
            <span class="text-sm text-gray-500">
              {{ slotProps.data.objective.info?.company }}
            </span>
          </div>
        </template>
      </Column>
      <Column field="contact" header="联系人" style="width: 120px" sortable />
      <Column field="callee" header="联系方式" style="width: 120px" sortable />
      <Column field="latestActivity" header="最近活动" style="width: 180px" sortable>
        <template #body="{ data }">
          <template v-if="getLatestActivity(data)">
            <div class="flex flex-row gap-4">
              <span class="text-sm text-gray-500">
                {{ formatRelativeTime(getLatestActivity(data).created) }}
              </span>
            </div>
          </template>
          <Tag v-else value="无记录" severity="secondary" class="text-xs" />
        </template>
      </Column>

      <!-- 新增操作列 -->
      <Column header="操作" style="width: 160px">
        <template #body="{ data }">
          <div class="flex gap-2">
            <Button label="查看" text severity="info" size="small" @click="viewTaskDetail(data)" />
            <ConfirmBtn
              v-tooltip="data.ext_id && data.id !== data.ext_id ? '需待源系统处理' : ''"
              :label="data.open ? '关闭' : '开启'"
              :severity="data.open ? 'danger' : 'success'"
              size="small"
              text
              :confirm-msg="`确定要${data.open ? '关闭' : '开启'}此任务吗？`"
              @confirm-act="toggleTaskStatus(data)"
            />
          </div>
        </template>
      </Column>
    </DataTable>

    <TaskDrawer
      v-model:visible="drawerVisible"
      :task-data="selectedTask"
      :objective="selectedTask?.objective"
      @need-refresh="loadData"
    />

    <FocusModeDrawer
      v-model:visible="focusDrawerVisible"
      :tasks="focusTasks"
      :initial-index="initialFocusIndex"
      @need-refresh="loadData"
    />
  </div>
</template>
