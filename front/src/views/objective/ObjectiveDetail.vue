<script setup>
import { ref, computed, onUnmounted, onMounted, inject } from 'vue'
import TaskDrawer from '@/components/TaskDrawer.vue'
import AddTask from '@/components/AddTask.vue'
import { pb } from '@/pocketbase/index'
import { useRoute } from 'vue-router'
import { formatRelativeTime } from '@/util/time'
import { useToast } from 'primevue/usetoast'
import ObjectiveInfo from '@/components/ObjectiveInfo.vue'

const toast = useToast()

// 注入父级提供的标题
const objectiveTitle = inject('objectiveTitle')
const showObjectiveDetail = inject('showObjectiveDetail')

import StatCard from '@/components/StatCard.vue'
import IconTaskList from '@/components/icons/IconTaskList.vue'
import IconSpin from '@/components/icons/IconSpin.vue'
import IconTrend from '@/components/icons/IconTrend.vue'
import IconPhone from '@/components/icons/IconPhone.vue'

const route = useRoute()
const object = ref()
const showTaskDialog = ref(false)
const showAddTask = ref(false)
const drawerVisible = ref(false)
const selectedTask = ref(null)

const handleTaskSelect = (event) => {
  selectedTask.value = event.data
  drawerVisible.value = true
}

async function loadData() {
  object.value = await pb.collection('objective').getOne(route.params.id, {
    expand: 'tasks, tasks.activity, tasks.own, tasks.activity.user'
  })
  objectiveTitle.value = object.value.title // 更新提供的标题

  if (selectedTask.value) {
    selectedTask.value = object.value?.expand?.tasks?.find((t) => t.id === selectedTask.value.id)
  }
}

onMounted(() => {
  loadData()
})

onUnmounted(() => {
  objectiveTitle.value = null
})

// 合并所有任务的 activity
const mergedActivities = computed(() => {
  return (
    object.value?.expand?.tasks
      ?.flatMap((t) => t.expand?.activity || [])
      .sort((a, b) => new Date(b.created) - new Date(a.created)) || []
  )
})

// 统计本周活动数
const weeklyActivities = computed(() => {
  const now = new Date()
  return mergedActivities.value.filter((a) => {
    const created = new Date(a.created)
    return now - created < 7 * 24 * 60 * 60 * 1000
  }).length
})

// 计算进行中任务数
const openTasksCount = computed(() => {
  return object.value?.expand?.tasks?.filter((t) => t.open).length || 0
})

// 计算接通率
const answerRate = computed(() => {
  const totalCalls = mergedActivities.value.filter((a) => a.isCall).length
  const answeredCalls = mergedActivities.value.filter((a) => a.isCall && a.record).length
  return totalCalls > 0 ? Math.round((answeredCalls / totalCalls) * 100) : 0
})

async function toogleTask(task, flag) {
  try {
    const data = await pb.collection('task').update(task.id, { open: flag })
    await loadData()
    const msg = data.message ? '.' + data.message : ''
    toast.add({
      severity: 'success',
      summary: '操作成功',
      detail: `任务已${flag ? '开启' : '关闭'}${msg}`,
      life: 3000
    })
  } catch (error) {
    console.log('eee!!', error)
    toast.add({
      severity: 'error',
      summary: '操作失败',
      detail: error.message,
      life: 3000
    })
  }
}
</script>

<template>
  <div v-if="object" class="flex flex-col gap-6">
    <!-- 统计卡片 -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <StatCard
        title="总任务数"
        :value="object.tasks?.length || 0"
        :icon="IconTaskList"
        class="text-blue-500"
      />
      <StatCard title="进行中" :value="openTasksCount" :icon="IconSpin" class="text-green-500" />
      <StatCard
        title="本周活动"
        :value="weeklyActivities"
        :icon="IconTrend"
        class="text-indigo-500"
      />
      <StatCard
        title="接通率"
        :value="answerRate"
        :suffix="'%'"
        :icon="IconPhone"
        class="text-orange-500"
      />
    </div>

    <!-- 关联任务 -->
    <div>
      <DataTable
        :value="object.expand?.tasks"
        class="p-datatable-sm"
        :paginator="true"
        :rows="10"
        scrollable
        scroll-height="600px"
        selection-mode="single"
        @row-click="handleTaskSelect"
      >
        <Column header="负责人" sort-field="expand.own.name">
          <template #body="{ data }">
            <Tag :value="data.expand?.own?.name || '未分配'" severity="info" />
          </template>
        </Column>
        <Column field="desc" header="任务简述" />
        <Column field="contact" header="客户" />
        <Column field="callee" header="联系方式" />
        <Column header="状态" style="width: 100px">
          <template #body="{ data }">
            <Tag
              :value="data.open ? '进行中' : '已关闭'"
              :severity="data.open ? 'success' : 'danger'"
              class="text-sm"
            />
          </template>
        </Column>
        <Column header="最近活动" style="width: 200px">
          <template #body="{ data }">
            <div v-if="data.expand?.activity?.[data.expand.activity.length - 1]" class="text-sm">
              <span class="text-gray-500">{{
                formatRelativeTime(data.expand.activity[data.expand.activity.length - 1].created)
              }}</span>
            </div>
          </template>
        </Column>
        <Column header="" style="width: 200px">
          <template #header>
            <div class="flex items-center justify-between w-full">
              <Button
                label="添加"
                icon="pi pi-plus"
                severity="success"
                size="small"
                @click="showAddTask = true"
              />
            </div>
          </template>
          <template #body="{ data }">
            <div class="flex gap-2">
              <ConfirmBtn
                v-if="data.open"
                v-tooltip="data.ext_id && data.id !== data.ext_id ? '需待源系统处理' : ''"
                label="关闭任务"
                severity="info"
                text
                confirm-msg="确定要关闭此任务吗？"
                @confirm-act="toogleTask(data, false)"
              />
              <ConfirmBtn
                v-else
                v-tooltip="data.ext_id && data.id !== data.ext_id ? '需待源系统处理' : ''"
                label="开启任务"
                severity="info"
                text
                confirm-msg="确定要重新开启此任务？"
                @confirm-act="toogleTask(data, true)"
              />
            </div>
          </template>
        </Column>
      </DataTable>
    </div>

    <ObjectiveInfo v-model:visible="showObjectiveDetail" :object="object" />

    <TaskDrawer
      v-model:visible="drawerVisible"
      :task-data="selectedTask"
      :objective="object"
      @need-refresh="loadData"
    />
    <AddTask v-model:visible="showAddTask" :objective="object" @task-created="loadData" />
    <Dialog v-model:visible="showTaskDialog" header="新建任务" :style="{ width: '600px' }">
      <div class="p-4">
        <p class="text-gray-500">任务表单组件需要根据实际业务需求实现</p>
      </div>
    </Dialog>
  </div>
</template>
