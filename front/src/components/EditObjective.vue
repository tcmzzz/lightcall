<script setup>
import { SchemaObjective } from '@/schema'
import { GenValidFn, GenSaveFn } from '@/valid'
import { useToast } from 'primevue/usetoast'
import { ref, watch } from 'vue'

const toast = useToast()
const visible = defineModel('visible', { type: Boolean, required: true })
const obj = defineModel('obj', { type: Object, default: () => ({ info: {} }) })
const emit = defineEmits(['saveSuc'])

const errors = ref({})
const fileUploader = ref(null)

const title = ref('')
watch(visible, (newV) => {
  if (newV) {
    errors.value = {}
    if (fileUploader.value) {
      fileUploader.value.clear()
    }
    title.value = obj.value.id ? '编辑目标' : '新建目标'
    if (!obj.value.info) obj.value.info = {}
  }
})

const valiObj = GenValidFn(SchemaObjective, obj, errors)
const saveObj = GenSaveFn(SchemaObjective, obj, errors, async (data) => {
  try {
    let record = data
    if (data.id) {
      // 编辑模式
      record = await pb.collection('objective').update(data.id, data)
    } else {
      // 创建模式
      data['open'] = true
      record = await pb.collection('objective').create(data)
    }

    // 如果有文件需要上传
    const filesToUpload = fileUploader.value.files
    if (filesToUpload && filesToUpload.length > 0) {
      const formData = new FormData()
      for (const file of filesToUpload) {
        formData.append('docs', file)
      }
      // 更新记录以上传文件
      await pb.collection('objective').update(record.id, formData)
    }

    toast.add({ severity: 'success', summary: '操作成功', life: 3000 })
    emit('saveSuc')
    visible.value = false
  } catch (error) {
    toast.add({ severity: 'error', summary: error.message, life: 3000 })
  }
})
</script>

<template>
  <Dialog v-model:visible="visible" modal :header="title" class="w-1/2">
    <div class="grid gap-4 p-4">
      <div class="flex items-center gap-4">
        <label class="font-semibold w-20">目标名称</label>
        <InputText
          v-model="obj.title"
          v-tooltip="errors.title"
          :invalid="!!errors.title"
          class="flex-auto"
          @update:model-value="valiObj('title')"
        />
      </div>

      <div class="flex flex-col gap-2">
        <div class="flex items-center gap-4">
          <label class="font-semibold w-20">公司名称</label>
          <InputText
            v-model="obj.info.company"
            v-tooltip="errors.info?.company"
            :invalid="!!errors.info?.company"
            class="flex-auto"
            @update:model-value="valiObj('info.company')"
          />
        </div>

        <div class="flex gap-4">
          <label class="font-semibold w-20">背景信息</label>
          <div class="flex-auto">
            <Textarea
              v-model="obj.info.background"
              v-tooltip="errors.info?.background"
              :invalid="!!errors.info?.background"
              class="w-full"
              rows="5"
              placeholder="支持Markdown格式：**粗体**、*斜体*、[链接](url)"
              @update:model-value="valiObj('info.background')"
            />
            <p class="text-xs text-gray-500 mt-1">
              支持Markdown格式：**粗体**、*斜体*、- 列表、[链接](url)
            </p>
          </div>
        </div>
      </div>

      <div class="flex flex-col gap-2 mt-4">
        <label class="font-semibold w-20">相关资料</label>
        <FileUpload
          ref="fileUploader"
          name="docs"
          :multiple="true"
          :custom-upload="true"
          :show-upload-button="false"
          :max-file-size="10485760"
          accept="image/*,.pdf"
        >
          <template #empty>
            <p>拖拽文件到这里 (支持pdf, png/jpg/webp).</p>
          </template>
        </FileUpload>
      </div>

      <div class="flex justify-end gap-2 mt-4">
        <Button label="取消" severity="secondary" @click="visible = false" />
        <Button label="保存" @click="saveObj" />
      </div>
    </div>
  </Dialog>
</template>
