<script setup>
import { useToast } from 'primevue/usetoast'

const toast = useToast()

function loadData() {
  pb.collection('outgw')
    .getFullList({ sort: '-created' })
    .then((data) => (gateways.value = data))
}

onMounted(() => {
  loadData()
})

const popGwData = ref({})
const gateways = ref([])
const visible = ref(false)

function popGwEdit(nv) {
  visible.value = true
  popGwData.value = JSON.parse(JSON.stringify(nv))
}

function delGw(rec) {
  pb.collection('outgw')
    .delete(rec.id)
    .then(() => {
      toast.add({
        severity: 'success',
        summary: '操作',
        detail: '删除成功',
        life: 3000
      })
      loadData()
    })
}

function toggleGatewayEnable(gw) {
  const updatedGw = { ...gw, enable: !gw.enable }
  pb.collection('outgw')
    .update(gw.id, updatedGw)
    .then(() => {
      toast.add({
        severity: 'success',
        summary: '操作',
        detail: updatedGw.enable ? '已启用' : '已停用',
        life: 3000
      })
      loadData()
    })
    .catch((error) => {
      toast.add({
        severity: 'error',
        summary: '操作失败',
        detail: error.message,
        life: 5000
      })
    })
}
function newGw() {
  popGwData.value = {
    name: '',
    protocol: '',
    addr: '',
    enable: true,
    options: {},
    transcaller: [],
    transcallee: []
  }
  visible.value = true
}
</script>
<template>
  <EditGateway v-model:visible="visible" v-model:gw="popGwData" @save-suc="loadData()" />
  <div class="flex">
    <Button class="ml-auto" text severity="info" label="添加" @click="newGw" />
  </div>
  <DataTable :value="gateways" paginator :rows="5" scrollable scroll-height="flex">
    <Column field="name" header="网关名称" />
    <Column field="protocol" header="协议" />
    <Column field="addr" header="地址" />
    <Column field="enable" header="启用" style="width: 8%">
      <template #body="slotProps">
        <i
          v-if="slotProps.data.enable"
          class="pi pi-check-circle text-green-500"
          style="font-size: 1.2rem"
        ></i>
        <i v-else class="pi pi-times-circle text-red-500" style="font-size: 1.2rem"></i>
      </template>
    </Column>
    <Column header="操作">
      <template #body="slotProps">
        <Button
          text
          :severity="slotProps.data.enable ? 'secondary' : 'success'"
          :label="slotProps.data.enable ? '停用' : '启用'"
          @click="toggleGatewayEnable(slotProps.data)"
        />
        <Button text severity="info" label="修改" @click="popGwEdit(slotProps.data)" />
        <ConfirmBtn
          text
          severity="warn"
          label="删除"
          confirm-msg="确定要删除吗？"
          @confirm-act="delGw(slotProps.data)"
        />
      </template>
    </Column>
  </DataTable>
</template>
