<script setup>
import { ref, onMounted, watch } from 'vue'
import { SchemaTask } from '@/schema'
import { GenValidFn, GenSaveFn } from '@/valid'
import { useToast } from 'primevue/usetoast'
import { pb } from '@/pocketbase/index'

const toast = useToast()
const emit = defineEmits(['task-created'])
const visible = defineModel('visible', { type: Boolean, required: true })
const props = defineProps({
  objective: { type: Object, required: true }
})

const errors = ref({})
const task = ref({
  desc: '',
  contact: '',
  callee: '',
  own: '',
  open: true
})

const userOptions = ref([])
const loading = ref(false)

watch(visible, async (newVal) => {
  if (newVal) {
    try {
      loading.value = true
      userOptions.value = await pb.collection('users').getFullList({
        fields: 'id,name',
        sort: 'name',
        $autoCancel: false
      })
    } catch (error) {
      toast.add({
        severity: 'error',
        summary: '加载失败',
        detail: '用户列表获取失败: ' + error.message,
        life: 3000
      })
    } finally {
      loading.value = false
    }
  }
})

const valiTask = GenValidFn(SchemaTask, task, errors)
const saveTask = GenSaveFn(SchemaTask, task, errors, async (taskData) => {
  try {
    const newTask = await pb.collection('task').create(taskData)
    await pb.collection('objective').update(props.objective.id, {
      'tasks+': newTask.id
    })

    toast.add({
      severity: 'success',
      summary: '任务添加成功',
      life: 3000
    })

    // 重置表单
    task.value = {
      desc: '',
      contact: '',
      callee: '',
      own: '',
      open: true
    }
    errors.value = {}
    visible.value = false
    emit('task-created')
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: error.message,
      life: 3000
    })
  }
})
</script>

<template>
  <Dialog v-model:visible="visible" modal style="width: 600px">
    <template #header>
      <h2 class="text-xl text-gray-900">
        新建任务 <span class="ml-3 text-gray-500"> {{ objective.title }} </span>
      </h2>
    </template>
    <div class="p-4 grid gap-4">
      <div class="flex items-center gap-4">
        <label class="font-semibold w-20">客户名称</label>
        <InputText
          v-model="task.contact"
          v-tooltip="errors.contact"
          :invalid="!!errors.contact"
          class="flex-grow"
          @update:model-value="valiTask('contact')"
        />
      </div>

      <div class="flex items-center gap-4">
        <label class="font-semibold w-20">联系方式</label>
        <InputText
          v-model="task.callee"
          v-tooltip="errors.callee"
          :invalid="!!errors.callee"
          class="flex-grow"
          @update:model-value="valiTask('callee')"
        />
      </div>

      <div class="flex items-center gap-4">
        <label class="font-semibold w-20">负责人</label>
        <Dropdown
          v-model="task.own"
          v-tooltip="errors.own"
          :options="userOptions"
          option-label="name"
          option-value="id"
          :loading="loading"
          :disabled="loading"
          :filter="true"
          placeholder="选择负责人"
          class="flex-grow"
          :invalid="!!errors.own"
          @update:model-value="valiTask('own')"
        >
          <template #empty>
            <div class="p-2 text-gray-500 text-sm">
              {{ loading ? '正在加载用户列表...' : '暂无可用用户' }}
            </div>
          </template>
          <template #option="slotProps">
            <div class="flex items-center gap-2">
              <i class="pi pi-user" />
              <span>{{ slotProps.option.name }}</span>
            </div>
          </template>
        </Dropdown>
      </div>

      <div class="flex flex-col gap-2">
        <div class="flex gap-4">
          <label class="font-semibold w-20">任务描述</label>
          <div class="flex-auto">
            <Textarea
              v-model="task.desc"
              v-tooltip="errors.desc"
              :invalid="!!errors.desc"
              class="w-full"
              rows="4"
              placeholder="支持Markdown格式：**粗体**、*斜体*、[链接](url)、- 列表"
              @update:model-value="valiTask('desc')"
            />
            <p class="text-xs text-gray-500 mt-1">
              支持Markdown格式：**粗体**、*斜体*、- 列表、[链接](url)
            </p>
          </div>
        </div>
      </div>

      <div class="flex justify-end gap-2">
        <Button label="取消" severity="secondary" @click="visible = false" />
        <Button label="保存" @click="saveTask" />
      </div>
    </div>
  </Dialog>
</template>
