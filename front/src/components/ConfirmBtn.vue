<script setup>
import { useConfirm } from 'primevue/useconfirm'

const confirm = useConfirm()
const props = defineProps({
  confirmMsg: {
    type: String,
    required: true
  }
})
const emit = defineEmits(['confirmAct'])

const handleClick = (event) => {
  confirm.require({
    target: event.currentTarget,
    message: props.confirmMsg,
    icon: 'pi pi-exclamation-triangle',
    rejectProps: { label: '取消', severity: 'secondary', outlined: true },
    acceptProps: { label: '确认' },
    accept: () => {
      emit('confirmAct')
    }
  })
}
</script>

<template>
  <Button @click="handleClick($event)" />
</template>
