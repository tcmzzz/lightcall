<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { pb } from '@/pocketbase/index'
import { useToast } from 'primevue/usetoast'
import { useUserStore } from '@/stores/user'

const toast = useToast()
const router = useRouter()
const userStore = useUserStore()

const users = ref([])
const visible = ref(false)
const currentUser = ref({})
const loading = ref(false)

// 加载用户数据
async function loadUsers() {
  try {
    loading.value = true
    users.value = await pb.collection('users').getFullList({
      sort: '-created'
      //fields: 'id,avatar,name,email,isAdmin,created,updated'
    })
    console.log(users.value)
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '加载失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}

// 初始化编辑表单
function initEdit(user = null) {
  currentUser.value = user
    ? { ...user, password: '' }
    : { name: '', email: '', password: '', isAdmin: false, verified: true, active: true }
  visible.value = true
}

// 保存用户
async function saveUser() {
  try {
    loading.value = true
    const userData = { ...currentUser.value }

    // 移除空密码字段
    if (!userData.password) delete userData.password

    if (userData.id) {
      if (userData.password) userData['passwordConfirm'] = userData.password
      await pb.collection('users').update(userData.id, userData)
    } else {
      userData['passwordConfirm'] = userData.password
      userData['emailVisibility'] = true
      userData['verified'] = true

      await pb.collection('users').create(userData)
    }

    toast.add({
      severity: 'success',
      summary: '操作成功',
      detail: `用户${userData.id ? '更新' : '创建'}成功`,
      life: 3000
    })
    loadUsers()
    visible.value = false
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '操作失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    loading.value = false
  }
}

// 切换用户状态
async function toggleActive(user) {
  try {
    const newActive = !user.active
    await pb.collection('users').update(user.id, { active: newActive })

    toast.add({
      severity: 'success',
      summary: '操作成功',
      detail: `用户已${newActive ? '启用' : '停用'}`,
      life: 3000
    })

    // 刷新用户列表
    loadUsers()
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '操作失败',
      detail: error.message,
      life: 3000
    })
  }
}

onMounted(() => {
  if (!userStore.userInfo?.isAdmin) {
    router.push('/')
    toast.add({
      severity: 'warn',
      summary: '权限不足',
      detail: '仅管理员可访问用户管理',
      life: 3000
    })
    return
  }
  loadUsers()
})
</script>

<template>
  <div class="flex justify-between mb-4">
    <h2 class="text-xl font-semibold">用户管理</h2>
    <Button label="添加用户" icon="pi pi-plus" @click="initEdit()" />
  </div>

  <DataTable :value="users" :loading="loading" paginator :rows="10">
    <Column field="name" header="姓名" sortable />
    <Column field="email" header="邮箱" sortable />
    <Column header="管理员" sortable>
      <template #body="{ data }">
        <Tag :severity="data.isAdmin ? 'success' : 'info'">
          {{ data.isAdmin ? '是' : '否' }}
        </Tag>
      </template>
    </Column>
    <Column header="状态" sortable>
      <template #body="{ data }">
        <Tag :severity="data.active ? 'success' : 'danger'">
          {{ data.active ? '启用' : '停用' }}
        </Tag>
      </template>
    </Column>
    <Column header="操作">
      <template #body="{ data }">
        <Button icon="pi pi-pencil" text @click="initEdit(data)" />
        <Button
          :icon="data.active ? 'pi pi-ban' : 'pi pi-check'"
          :severity="data.active ? 'danger' : 'success'"
          text
          @click="toggleActive(data)"
        />
      </template>
    </Column>
  </DataTable>

  <!-- 用户编辑对话框 -->
  <Dialog v-model:visible="visible" modal header="用户信息" :style="{ width: '450px' }">
    <div class="ml-12 flex flex-col gap-3">
      <div class="flex items-center gap-5">
        <label class="font-semibold">姓名</label>
        <InputText v-model="currentUser.name" required />
      </div>

      <div class="flex items-center gap-5">
        <label class="font-semibold">邮箱</label>
        <InputText v-model="currentUser.email" type="email" required />
      </div>

      <div class="flex items-center gap-5">
        <label class="font-semibold">密码</label>
        <div class="w-12">
          <Password v-model="currentUser.password" toggle-mask :feedback="false" />
        </div>
      </div>

      <div class="flex items-center gap-5">
        <label class="font-semibold">管理员权限</label>
        <Checkbox v-model="currentUser.isAdmin" :binary="true" />
      </div>
      <div class="flex items-center gap-5">
        <label class="font-semibold">启用状态</label>
        <InputSwitch v-model="currentUser.active" />
      </div>
    </div>

    <template #footer>
      <Button label="取消" severity="secondary" @click="visible = false" />
      <Button label="保存" :loading="loading" @click="saveUser" />
    </template>
  </Dialog>
</template>
