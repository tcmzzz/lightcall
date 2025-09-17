<script setup>
import { ref, onMounted, watch } from 'vue'
import { pb } from '@/pocketbase/index'
import { useToast } from 'primevue/usetoast'
import ConfirmBtn from '@/components/ConfirmBtn.vue'
import EditObjective from '@/components/EditObjective.vue'
import ObjectiveInfo from '@/components/ObjectiveInfo.vue'

const showEdit = ref(false)
const editObj = ref({})
const showInfoDialog = ref(false)
const selectedObject = ref(null)

function showObjectInfo(obj) {
  selectedObject.value = obj
  showInfoDialog.value = true
}

const statusFilter = ref(true)
const statusOptions = [
  { label: '全部', value: null },
  { label: '进行中', value: true },
  { label: '已归档', value: false }
]

function newObjective() {
  editObj.value = { info: {} }
  showEdit.value = true
}

async function handleSave() {
  await loadData()
}

const hasRecentActivity = (objective) => {
  const now = new Date()
  return (
    objective.expand?.tasks?.some((task) =>
      task.expand?.activity?.some((activity) => {
        const created = new Date(activity.created)
        return now - created < 24 * 60 * 60 * 1000
      })
    ) ?? false
  )
}
const toast = useToast()

const objects = ref([])

async function loadData() {
  try {
    let filter = ''
    if (statusFilter.value !== null) {
      filter = `open = ${statusFilter.value}`
    }

    objects.value = await pb.collection('objective').getFullList({
      sort: '-created',
      expand: 'tasks, tasks.activity', // 修改expand参数
      filter: filter
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '加载失败',
      detail: error.message,
      life: 3000
    })
  }
}

watch(statusFilter, () => {
  loadData()
})

onMounted(() => {
  loadData()
})

function deleteObject(obj) {
  pb.collection('objective')
    .delete(obj.id)
    .then(() => {
      toast.add({
        severity: 'success',
        summary: '删除成功',
        detail: '目标已永久删除',
        life: 3000
      })
      loadData()
    })
}

function openObject(obj) {
  pb.collection('objective')
    .update(obj.id, { open: !obj.open })
    .then((data) => {
      const msg = data.message ? '.' + data.message : ''
      toast.add({
        severity: 'success',
        summary: '操作成功',
        detail: obj.open ? '目标已归档' + msg : '已取消归档' + msg,
        life: 3000
      })
      loadData()
    })
}
function truncate(text, length) {
  if (text && text.length > length) {
    return text.substring(0, length) + '...'
  }
  return text
}
</script>

<template>
  <div class="flex items-center mb-4 mt-5">
    <div class="flex-1">
      <label class="mr-2 text-gray-500">状态:</label>
      <Dropdown
        v-model="statusFilter"
        :options="statusOptions"
        option-label="label"
        option-value="value"
        placeholder="筛选状态"
        class="w-48"
      />
    </div>
    <Button class="ml-auto" text severity="info" label="添加" @click="newObjective" />
    <EditObjective v-model:visible="showEdit" v-model:obj="editObj" @save-suc="handleSave" />
  </div>

  <DataTable
    :value="objects"
    paginator
    :rows="10"
    scrollable
    scroll-height="flex"
    class="p-datatable-sm"
  >
    <Column field="title" header="目标名称" :width="300">
      <template #body="{ data }">
        <div class="flex items-center gap-2">
          <Tag
            :value="data.open ? '进行中' : '已归档'"
            :severity="data.open ? 'success' : 'danger'"
          />
          {{ data.title }}
          <Tag v-if="data.ext_id && data.id !== data.ext_id" value="同步" severity="info" />
        </div>
      </template>
    </Column>

    <Column header="关联任务" :width="200">
      <template #body="{ data }">
        <div class="flex gap-2">
          <span class="text-sm text-gray-600"> 总数: {{ data.tasks?.length || 0 }} </span>
          <span class="text-sm text-green-600">
            进行中: {{ data.expand?.tasks?.filter((t) => t.open).length || 0 }}
          </span>
        </div>
      </template>
    </Column>

    <Column header="公司信息" :width="300">
      <template #body="{ data }">
        <div class="flex items-start justify-between">
          <div class="flex flex-col flex-grow overflow-hidden">
            <span v-tooltip.top="data.info?.company" class="font-medium">
              {{ truncate(data.info?.company, 50) }}
            </span>
          </div>
          <Button
            label="信息详情"
            severity="secondary"
            text
            size="small"
            class="mr-auto flex-shrink-0"
            @click="showObjectInfo(data)"
          />
        </div>
      </template>
    </Column>

    <Column header="近期活动" style="width: 120px">
      <template #body="{ data }">
        <Tag
          :value="hasRecentActivity(data) ? '有' : '无'"
          :severity="hasRecentActivity(data) ? 'success' : 'secondary'"
          class="text-sm"
        />
      </template>
    </Column>

    <Column header="操作" style="width: 240px">
      <template #body="{ data }">
        <div class="flex gap-3">
          <Button
            v-tooltip="'查看详情'"
            label="查看"
            severity="info"
            class="p-button-sm"
            text
            @click="$router.push(`/objective/detail/${data.id}`)"
          />
          <ConfirmBtn
            v-tooltip="data.ext_id && data.id !== data.ext_id ? '需待源系统处理' : ''"
            :label="data.open ? '归档' : '取消归档'"
            :severity="data.open ? 'info' : 'secondary'"
            class="p-button-sm"
            text
            :confirm-msg="`确定要${data.open ? '归档' : '取消归档'}该目标吗？`"
            @confirm-act="openObject(data)"
          />
        </div>
      </template>
    </Column>
  </DataTable>
  <ObjectiveInfo v-if="selectedObject" v-model:visible="showInfoDialog" :object="selectedObject" />
</template>
