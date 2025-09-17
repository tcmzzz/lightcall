<script setup>
import { useToast } from 'primevue/usetoast'
const toast = useToast()

const numbers = ref([])
const numEditObj = ref({})
const numEditVisible = ref(false)

async function loadData() {
  numbers.value = await pb.collection('number').getFullList({
    sort: '-created',
    expand: 'outgw'
  })
}

onMounted(async () => {
  loadData()
})

function popNumEdit(num) {
  numEditObj.value = JSON.parse(JSON.stringify(num))
  numEditVisible.value = true
}

function delNum(obj) {
  pb.collection('number')
    .delete(obj.id)
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

function toggleNumberEnable(obj) {
  const updatedObj = { ...obj, enable: !obj.enable }
  pb.collection('number')
    .update(obj.id, updatedObj)
    .then(() => {
      toast.add({
        severity: 'success',
        summary: '操作',
        detail: updatedObj.enable ? '已启用' : '已停用',
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
function newNum() {
  numEditObj.value = {
    enable: true,
    tag: {}
  }
  numEditVisible.value = true
}
</script>
<template>
  <div>
    <EditNumber v-model:visible="numEditVisible" v-model:obj="numEditObj" @save-suc="loadData()" />
    <div class="flex">
      <Button class="ml-auto" text severity="info" label="添加" @click="newNum" />
    </div>
    <DataTable :value="numbers" paginator :rows="5" scrollable scroll-height="flex">
      <Column field="number" header="号码" />
      <Column field="expand.outgw.name" header="网关" />
      <Column header="归属地(省市)">
        <template #body="slotProps">
          <Tag severity="info" :value="slotProps.data.tag.province" />
          <Tag class="ml-3" severity="info" :value="slotProps.data.tag.city" />
        </template>
      </Column>
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
            @click="toggleNumberEnable(slotProps.data)"
          />
          <Button text severity="info" label="修改" @click="popNumEdit(slotProps.data)" />
          <ConfirmBtn
            text
            severity="warn"
            label="删除"
            confirm-msg="确定要删除吗？"
            @confirm-act="delNum(slotProps.data)"
          />
        </template>
      </Column>
    </DataTable>
  </div>
</template>
