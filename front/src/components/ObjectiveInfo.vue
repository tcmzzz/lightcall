<script setup>
import { formatAbsoluteTime } from '@/util/time'
import { renderMarkdown } from '@/util/markdown'
import { pb } from '@/pocketbase/index'
import { ref, watch, computed, nextTick } from 'vue'
import { useToast } from 'primevue/usetoast'
import Galleria from 'primevue/galleria'
import Image from 'primevue/image'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const toast = useToast()

const props = defineProps({
  object: {
    type: Object,
    required: true
  },
  visible: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['update:visible', 'updated'])

const objective = ref(null)
const imageDocs = ref([])
const pdfDocs = ref([])
const isEditingBackground = ref(false)
const editableBackground = ref('')

// 开始编辑
function startEditBackground() {
  editableBackground.value = objective.value.info.background
  isEditingBackground.value = true
}

// 取消编辑
function cancelEditBackground() {
  isEditingBackground.value = false
}

// 保存背景信息
async function updateBackground() {
  if (!objective.value) return
  const newInfo = { ...objective.value.info, background: editableBackground.value }
  try {
    await pb.collection('objective').update(objective.value.id, { info: newInfo })
    await refreshObjective() // 使用现有函数刷新数据
    isEditingBackground.value = false
    toast.add({ severity: 'success', summary: '成功', detail: '背景信息已更新', life: 3000 })
  } catch (error) {
    toast.add({ severity: 'error', summary: '错误', detail: '更新失败', life: 3000 })
  }
}

watch(
  () => props.object,
  async (newVal) => {
    objective.value = newVal
    await fetchDocURL()
    await nextTick()
  },
  { immediate: true }
)

async function fetchDocURL() {
  const token = await pb.files.getToken()

  imageDocs.value =
    objective.value.docs
      ?.filter((doc) => isImage(doc))
      .map((doc) => {
        return {
          name: doc,
          url: pb.files.getURL(objective.value, doc, { token: token })
        }
      }) || []

  pdfDocs.value =
    objective.value.docs
      ?.filter((doc) => !isImage(doc))
      .map((doc) => {
        return {
          name: doc,
          url: pb.files.getURL(objective.value, doc, { token: token })
        }
      }) || []

  console.log(imageDocs.value, pdfDocs.value)
}

function closeModal() {
  emit('update:visible', false)
}

const isImage = (docName) => {
  if (!docName) return false
  const extension = docName.split('.').pop().toLowerCase()
  return ['png', 'jpg', 'jpeg', 'gif', 'webp'].includes(extension)
}

const deleteDoc = async (doc) => {
  if (!objective.value) return
  const newDocs = objective.value.docs.filter((d) => d !== doc.name)
  try {
    await pb.collection('objective').update(objective.value.id, { docs: newDocs })
    await refreshObjective()
    toast.add({ severity: 'success', summary: '成功', detail: '文件删除成功', life: 3000 })
  } catch (error) {
    toast.add({ severity: 'error', summary: '错误', detail: '文件删除失败', life: 3000 })
  }
}

const refreshObjective = async () => {
  if (!props.object) return
  objective.value = await pb.collection('objective').getOne(props.object.id)
  emit('updated')
}

const myUploader = async (event) => {
  const files = event.files
  if (!files || files.length === 0) {
    return
  }

  const formData = new FormData()
  if (objective.value.docs) {
    for (const doc of objective.value.docs) {
      formData.append('docs', doc)
    }
  }

  for (const file of files) {
    formData.append('docs', file)
  }

  try {
    await pb.collection('objective').update(objective.value.id, formData)
    await refreshObjective()
    toast.add({ severity: 'success', summary: '成功', detail: '文件上传成功', life: 3000 })
  } catch (error) {
    toast.add({ severity: 'error', summary: '错误', detail: '文件上传失败', life: 3000 })
  }
}
</script>

<template>
  <Dialog
    :visible="visible"
    header="目标详情"
    class="w-200 h-full"
    :style="{ height: '100vh' }"
    :closeOnEscape="true"
    @update:visible="closeModal"
  >
    <template #header>
      <h2 class="text-xl text-gray-900">
        <span class="text-gray-500">目标详情: </span> {{ objective.title }}
      </h2>
    </template>
    <div
      v-if="objective"
      class="p-4 space-y-4"
      style="max-height: calc(100vh - 100px); overflow-y: auto"
    >
      <div class="flex items-center justify-between text-sm">
        <div class="flex items-center gap-2">
          <span class="font-medium text-gray-500">公司名称:</span>
          <span class="text-gray-900">{{ objective.info?.company }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="font-medium text-gray-500">创建时间:</span>
          <span class="text-gray-900">{{ formatAbsoluteTime(objective.created) }}</span>
        </div>
      </div>
      <div>
        <dt class="text-sm font-medium text-gray-500 mb-1">相关资料</dt>
        <div
          v-if="!objective.docs || objective.docs.length === 0"
          class="flex items-center justify-center py-8"
        >
          暂无资料
        </div>
        <div v-else>
          <div v-if="imageDocs.length > 0" class="mb-4">
            <h3 class="font-semibold mb-2">图片</h3>
            <div class="flex justify-center">
              <Galleria
                :value="imageDocs"
                :num-visible="5"
                :circular="true"
                container-style="max-width: 640px"
                :show-thumbnails="false"
                :show-indicators="true"
                :show-item-navigators="true"
                :show-item-navigators-on-hover="true"
              >
                <template #item="slotProps">
                  <Image :src="slotProps.item.url" :alt="slotProps.item.name" width="640" preview />
                </template>
                <template #thumbnail="slotProps">
                  <img
                    :src="slotProps.item.url"
                    :alt="slotProps.item.name"
                    style="display: block"
                  />
                </template>
              </Galleria>
            </div>
            <div v-if="userStore.userInfo?.isAdmin">
              <ul class="space-y-2 mt-2">
                <li
                  v-for="doc in imageDocs"
                  :key="doc.name"
                  class="flex items-center justify-between"
                >
                  <div class="flex items-center">
                    - <img :src="doc.url" :alt="doc.name" class="w-10 h-10 object-cover mr-2" /> -
                    <span class="text-gray-900">{{ doc.name }}</span> -
                  </div>
                  <Button
                    v-if="userStore.userInfo?.isAdmin"
                    icon="pi pi-trash"
                    class="p-button-rounded p-button-danger p-button-text"
                    @click="deleteDoc(doc)"
                  />
                </li>
              </ul>
            </div>
          </div>

          <div v-if="pdfDocs.length > 0">
            <h3 class="font-semibold mb-2">文档</h3>
            <ul class="space-y-2">
              <li v-for="doc in pdfDocs" :key="doc.name" class="flex items-center justify-between">
                <a :href="doc.url" target="_blank" class="text-blue-500 hover:underline">{{
                  doc.name
                }}</a>
                <Button
                  v-if="userStore.userInfo?.isAdmin"
                  icon="pi pi-trash"
                  class="p-button-rounded p-button-danger p-button-text"
                  @click="deleteDoc(doc)"
                />
              </li>
            </ul>
          </div>
        </div>
      </div>
      <div v-if="userStore.userInfo?.isAdmin">
        <FileUpload
          name="docs"
          :multiple="true"
          :custom-upload="true"
          :show-upload-button="true"
          upload-label="上传"
          choose-label="选择文件"
          cancel-label="取消"
          :max-file-size="10485760"
          @uploader="myUploader"
        >
          <template #empty>
            <p>拖拽文件到这里来上传(支持pdf文档, 图片png/jpg/webp).</p>
          </template>
        </FileUpload>
      </div>
      <div>
        <div class="flex items-center justify-between mb-1">
          <dt class="text-sm font-medium text-gray-500">背景信息</dt>
          <Button
            v-if="userStore.userInfo?.isAdmin && !isEditingBackground"
            icon="pi pi-pencil"
            class="p-button-rounded p-button-text mr-auto"
            @click="startEditBackground"
          />
        </div>
        <template v-if="isEditingBackground">
          <Textarea v-model="editableBackground" class="w-full" rows="10" />
          <div class="flex justify-end gap-2 mt-2">
            <Button label="取消" severity="secondary" @click="cancelEditBackground" />
            <Button label="保存" @click="updateBackground" />
          </div>
        </template>
        <template v-else>
          <!-- eslint-disable vue/no-v-html -->
          <div class="bg-gray-50 p-4 rounded">
            <article
              class="prose prose-sm prose-gray max-w-full"
              v-html="renderMarkdown(objective.info?.background)"
            ></article>
          </div>
          <!-- eslint-enable vue/no-v-html -->
        </template>
      </div>
    </div>
  </Dialog>
</template>
