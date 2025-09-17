<script setup>
import { SchemaOutGwTrans } from '@/schema'
import { GenSaveFn } from '@/valid'
const visible = defineModel('visible', { type: Boolean, required: true })
const vm = defineModel('vm', { type: Object, required: true })

const transTps = [
  {
    type: 'prefix',
    name: '增加前缀',
    param: ['前缀']
  },
  {
    type: 'suffix',
    name: '增加后缀',
    param: ['后缀']
  },
  {
    type: 'replace',
    name: '替换',
    param: ['正则', '替换内容']
  }
]
const errors = ref()
const selected = ref({})
const addObj = ref({
  type: '',
  param: []
})

function init() {
  selected.value = transTps[0]
  addObj.value = {
    type: '',
    param: []
  }
  addObj.value.type = selected.value.type
  addObj.value.param = new Array(selected.value.param.length)
  errors.value = {}
  //errors.value.param = []
}

watch(visible, (newV) => {
  if (newV == true) {
    init()
  }
})

watch(selected, () => {
  addObj.value.type = selected.value.type
  addObj.value.param = new Array(selected.value.param.length)
  errors.value.param = []
})

const save = GenSaveFn(SchemaOutGwTrans, addObj, errors, (obj) => {
  if (vm.value != null) {
    vm.value.push(obj)
  } else {
    console.warn('can not reach here, vm should be array')
  }
  visible.value = false
})
</script>

<template>
  <Dialog v-model:visible="visible" modal header="Edit Profile" :style="{ width: '25rem' }">
    <span class="text-surface-500 dark:text-surface-400 block mb-8" />
    <div class="flex items-center gap-4 mb-4">
      <label for="username" class="font-semibold w-24">类型</label>
      <Select
        v-model="selected"
        :options="transTps"
        option-label="name"
        placeholder="请选择类型"
        :invalid="errors.type != null"
        class="w-full md:w-56"
      />
    </div>
    <div v-if="selected && selected.name != ''">
      <div v-for="(item, idx) in selected.param" :key="item" class="flex items-center gap-4 mb-2">
        <label class="font-semibold w-24">{{ item }}</label>
        <InputText
          v-model="addObj.param[idx]"
          v-tooltip="errors.param && errors.param[idx]"
          class="flex-auto"
          :invalid="errors.param && errors.param[idx] != null"
        />
      </div>
    </div>
    {{ errors }}
    <div class="flex justify-end gap-2 mt-8">
      <Button type="button" label="Cancel" severity="secondary" @click="visible = false" />
      <Button type="button" label="Save" @click="save" />
    </div>
  </Dialog>
</template>
