<script setup>
import { SchemaNumber } from '@/schema'
import { AreaObj } from '@/schema/area'
import { GenValidFn, GenSaveFn } from '@/valid'

import { useToast } from 'primevue/usetoast'

const toast = useToast()

const visible = defineModel('visible', { type: Boolean, required: true })
const obj = defineModel('obj', { type: Object, required: true })
const emit = defineEmits(['saveSuc'])

const errors = ref({})
const outgwOptioins = ref([])

const title = ref('')
watch(visible, async (newV) => {
  if (newV == true) {
    if (obj.value.id) {
      title.value = '编辑号码'
    } else {
      title.value = '新增号码'
    }

    errors.value = {}
    updateCity()
    outgwOptioins.value = await pb.collection('outgw').getFullList({ sort: '-created' })
  }
})

const cityOptions = ref()
const valiNumber = GenValidFn(SchemaNumber, obj, errors)
const saveNumber = GenSaveFn(SchemaNumber, obj, errors, async (obj) => {
  if (obj.id) {
    await pb.collection('number').update(obj.id, obj)
    toast.add({
      severity: 'success',
      summary: '操作',
      detail: '修改成功',
      life: 3000
    })
  } else {
    await pb.collection('number').create(obj)
    toast.add({
      severity: 'success',
      summary: '操作',
      detail: '添加成功',
      life: 3000
    })
  }

  emit('saveSuc')
  visible.value = false
})

function updateCity() {
  const found = AreaObj.find((elm) => {
    return elm.name == obj.value.tag.province
  })
  cityOptions.value = found?.cityList ? found.cityList : []
}
</script>

<template>
  <Dialog v-model:visible="visible" modal :header="title" style="width: 500px">
    <span class="text-surface-500 dark:text-surface-400 block mb-8" />

    <div class="flex items-center gap-4 mb-4">
      <label for="number" class="font-semibold w-24">号码</label>
      <InputText
        v-model="obj.number"
        v-tooltip="errors.number"
        :disabled="obj.id != null"
        class="w-full md:w-56"
        :invalid="errors.number != null"
        @update:model-value="valiNumber('number')"
      />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="outgw" class="font-semibold w-24">网关</label>

      <Select
        v-model="obj.outgw"
        :options="outgwOptioins"
        option-label="name"
        option-value="id"
        placeholder="选择一个出口网关"
        class="w-full md:w-56"
        :invalid="errors.outgw != null"
        @update:model-value="valiNumber('outgw')"
      />
      <span class="text-red-500 text-xs">{{ errors.outgw }}</span>
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="tag.province" class="font-semibold w-24">归属地(省)</label>
      <Select
        v-model="obj.tag.province"
        editable
        :options="AreaObj"
        option-label="name"
        option-value="name"
        placeholder="选择一个省份"
        class="w-full md:w-56"
        @update:model-value="updateCity"
      />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="tag.city" class="font-semibold w-24">归属地(市)</label>
      <Select
        v-model="obj.tag.city"
        editable
        :options="cityOptions"
        option-label="name"
        option-value="name"
        placeholder="选择一个城市"
        class="w-full md:w-56"
      />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="enable" class="font-semibold w-24">启用</label>
      <Checkbox v-model="obj.enable" :binary="true" />
    </div>
    <div class="flex justify-end gap-2">
      {{ errors }}
      <Button type="button" label="Cancel" severity="secondary" @click="visible = false" />
      <Button type="button" label="Save" @click="saveNumber" />
      <!--<ConfirmBtn label="Save" confirm-msg="确定要修改吗？" @confirm-act="popGwSave"></ConfirmBtn>-->
    </div>
  </Dialog>
</template>
