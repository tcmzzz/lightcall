<script setup>
import { SchemaOutGw } from '@/schema'
import { GenValidFn, GenSaveFn } from '@/valid'
import { useToast } from 'primevue/usetoast'

const toast = useToast()

const visible = defineModel('visible', { type: Boolean, required: true })
const gw = defineModel('gw', { type: Object, required: true })
const emit = defineEmits(['saveSuc'])

const errors = ref({})
const showEditTrans = ref(false)
const transVm = ref({})

const title = ref('')

watch(visible, (newV) => {
  if (newV == true) {
    errors.value = {}
    if (gw.value.id) {
      title.value = '编辑出口网关'
    } else {
      title.value = '新建出口网关'
    }
  }
})

const valiOutGw = GenValidFn(SchemaOutGw, gw, errors)
const saveOutGw = GenSaveFn(SchemaOutGw, gw, errors, async (obj) => {
  if (obj.id) {
    await pb.collection('outgw').update(obj.id, obj)
    toast.add({
      severity: 'success',
      summary: '操作',
      detail: '修改成功',
      life: 3000
    })
  } else {
    await pb.collection('outgw').create(obj)
    toast.add({
      severity: 'success',
      summary: '操作',
      detail: '创建成功',
      life: 3000
    })
  }
  emit('saveSuc')
  visible.value = false
})

function fmtTrans(rule) {
  if (rule.type === 'prefix') {
    return '增加前缀"' + (rule.param ? rule.param[0] : '') + '"'
  }

  if (rule.type === 'replace') {
    return (
      '替换/' + (rule.param ? rule.param[0] : '') + '/为"' + (rule.param ? rule.param[1] : '') + '"'
    )
  }
  if (rule.type === 'suffix') {
    return '增加后缀"' + (rule.param ? rule.param[0] : '') + '"'
  }

  return '缺少定义'
}
</script>

<template>
  <EditGatewayTrans v-model:visible="showEditTrans" v-model:vm="transVm" />
  <Dialog v-model:visible="visible" modal :header="title" class="w-2/5">
    <span class="text-surface-500 dark:text-surface-400 block mb-8" />
    <div class="flex items-center gap-4 mb-4">
      <label for="name" class="font-semibold w-24">网关名称</label>
      <InputText
        v-model="gw.name"
        v-tooltip="errors.name"
        class="flex-auto"
        :invalid="errors.name != null"
        @update:model-value="valiOutGw('name')"
      />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="protocol" class="font-semibold w-24">协议</label>
      <InputText
        v-model="gw.protocol"
        v-tooltip="errors.protocol"
        class="flex-auto"
        :invalid="errors.protocol != null"
        @update:model-value="valiOutGw('protocol')"
      />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="addr" class="font-semibold w-24">地址</label>
      <InputText
        v-model="gw.addr"
        v-tooltip="errors.addr"
        class="flex-auto"
        :invalid="errors.addr != null"
        @update:model-value="valiOutGw('addr')"
      />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="enable" class="font-semibold w-24">启用</label>
      <Checkbox v-model="gw.enable" :binary="true" />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="registry" class="font-semibold w-24">是否注册</label>
      <Checkbox v-model="gw.options.registry" :binary="true" />
      <label for="password" class="font-semibold text-center w-24">密码</label>
      <Password v-model="gw.options.password" :feedback="false" class="flex-auto" />
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="transcaller" class="font-semibold w-24">主叫号码变换</label>
      <div>
        <Tag
          v-for="(item, idx) in gw.transcaller"
          :key="item.type"
          v-tooltip="
            errors.transcaller && errors.transcaller[idx]
              ? errors.transcaller[idx].param.filter((i) => i != null).toString()
              : ''
          "
          :severity="errors.transcaller && errors.transcaller[idx] ? 'danger' : 'secondary'"
        >
          {{ fmtTrans(item) }}
          <span
            class="pi pi-times-circle"
            @click="(gw.transcaller.splice(idx, 1), valiOutGw('transcaller'))"
          />
        </Tag>
        <Chip icon="pi pi-plus" @click="((showEditTrans = true), (transVm = gw.transcaller))" />
      </div>
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="transcallee" class="font-semibold w-24">被叫号码变换</label>
      <div>
        <Tag
          v-for="(item, idx) in gw.transcallee"
          :key="item.type + item.param.length"
          v-tooltip="
            errors.transcallee && errors.transcallee[idx]
              ? errors.transcallee[idx].param.filter((i) => i != null).toString()
              : ''
          "
          :severity="errors.transcallee && errors.transcallee[idx] ? 'danger' : 'secondary'"
        >
          {{ fmtTrans(item) }}
          <span
            class="pi pi-times-circle"
            @click="(gw.transcallee.splice(idx, 1), valiOutGw('transcaller'))"
          />
        </Tag>

        <Chip icon="pi pi-plus" @click="((showEditTrans = true), (transVm = gw.transcallee))" />
      </div>
    </div>
    <div class="flex justify-end gap-2">
      {{ errors }}
      <Button type="button" label="Cancel" severity="secondary" @click="visible = false" />
      <Button type="button" label="Save" @click="saveOutGw" />
      <!--<ConfirmBtn label="Save" confirm-msg="确定要修改吗？" @confirm-act="popGwSave"></ConfirmBtn>-->
    </div>
  </Dialog>
</template>
