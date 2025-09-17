<script setup>
import { ref, onMounted, watch } from 'vue'
import { useToast } from 'primevue/usetoast'
import { SchemaConfigDial, SchemaConfigPrivacy, SchemaConfigCloud, SchemaConfigIce } from '@/schema'
import { GenValidFn, GenSaveFn } from '@/valid'

const toast = useToast()

// ICE config editing
const editingIceConfig = ref('')

// 配置数据 - 每个配置独立变量
const dialConfig = ref({ caller: { affinity: false } })
const privacyConfig = ref({ hideNumber: false })
const cloudConfig = ref({
  addr: '',
  appid: '',
  secret: '',
  lifecycle: { precall: { blacklist: false, flashCard: false } }
})
const iceConfig = ref([])

// 错误变量 - 每个配置独立
const dialErrors = ref({})
const privacyErrors = ref({})
const cloudErrors = ref({})
const iceErrors = ref({})

// 保存状态
const savingStates = ref({
  dial: false,
  privacy: false,
  cloud: false,
  ice: false,
  callSettings: false
})

// 加载配置
async function loadConfigs() {
  try {
    const records = await pb.collection('config').getFullList()
    records.forEach((record) => {
      if (record.name === 'dial') dialConfig.value = record.value
      if (record.name === 'privacy') privacyConfig.value = record.value
      if (record.name === 'cloud') cloudConfig.value = record.value
      if (record.name === 'ice_servers') {
        iceConfig.value = record.value
        editingIceConfig.value = JSON.stringify(record.value, null, 2)
      }
    })
  } catch (error) {
    toast.add({
      severity: 'error',
      summary: '加载配置失败',
      detail: error.message,
      life: 3000
    })
  }
}

// 验证函数 - 每个配置独立
const validateDial = GenValidFn(SchemaConfigDial, dialConfig, dialErrors)
const validatePrivacy = GenValidFn(SchemaConfigPrivacy, privacyConfig, privacyErrors)
const validateCloud = GenValidFn(SchemaConfigCloud, cloudConfig, cloudErrors)

// 保存函数 - 每个配置独立
const saveDial = GenSaveFn(SchemaConfigDial, dialConfig, dialErrors, async (validData) => {
  await saveConfig('dial', validData, 'dial', false)
})

const savePrivacy = GenSaveFn(
  SchemaConfigPrivacy,
  privacyConfig,
  privacyErrors,
  async (validData) => {
    await saveConfig('privacy', validData, 'privacy', false)
  }
)

const saveCloud = GenSaveFn(SchemaConfigCloud, cloudConfig, cloudErrors, async (validData) => {
  await saveConfig('cloud', validData, 'cloud')
})

async function saveConfig(name, value, type, showToast = true) {
  savingStates.value[type] = true
  try {
    const record = await pb.collection('config').getFirstListItem(`name="${name}"`)
    console.log(record)
    await pb.collection('config').update(record.id, { value })
    if (showToast) {
      toast.add({
        severity: 'success',
        summary: '保存成功',
        detail: `配置已更新`,
        life: 3000
      })
    }
  } catch (error) {
    if (showToast) {
      toast.add({
        severity: 'error',
        summary: '保存失败',
        detail: error.message,
        life: 3000
      })
    }
    throw error
  } finally {
    savingStates.value[type] = false
  }
}

// 合并保存呼叫设置和隐私设置
async function saveCallSettings() {
  // 验证两个配置
  savingStates.value.callSettings = true
  try {
    await saveDial()
    await savePrivacy()

    // 只显示一个成功的toast
    toast.add({
      severity: 'success',
      summary: '保存成功',
      detail: '呼叫设置已更新',
      life: 3000
    })
  } catch (error) {
    // 如果任何一个保存失败，显示错误toast
    toast.add({
      severity: 'error',
      summary: '保存失败',
      detail: error.message,
      life: 3000
    })
  } finally {
    savingStates.value.callSettings = false
  }
}

async function handleIceConfig(isSave) {
  console.log(isSave)
  try {
    const parsed = JSON.parse(editingIceConfig.value || '[]')
    iceConfig.value = SchemaConfigIce.cast(parsed)
    await SchemaConfigIce.validate(iceConfig.value, { abortEarly: false })

    iceErrors.value = {}
    if (isSave) {
      await saveConfig('ice_servers', iceConfig.value, 'ice')
      await loadConfigs()
    }
  } catch (error) {
    if (error instanceof SyntaxError) {
      iceErrors.value = { general: '请输入有效的JSON格式' }
    } else {
      if (error.inner) {
        const errorObj = {}
        error.inner.forEach((err) => {
          errorObj[err.path] = err.message
        })
        iceErrors.value = errorObj
      } else {
        iceErrors.value = { general: error.message }
      }
      console.log(iceErrors)
    }
  }
}

onMounted(() => {
  loadConfigs()
})
</script>

<template>
  <!-- 响应式两栏布局 -->
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- 左栏：基础设置 -->
    <div class="space-y-4">
      <!-- 呼叫设置 -->
      <Card>
        <template #title>
          <div class="flex items-center gap-2">
            <i class="pi pi-phone text-blue-500"></i>
            <span>呼叫设置</span>
          </div>
        </template>
        <template #content>
          <div class="space-y-6">
            <!-- 亲和性呼叫 -->
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <label class="text-sm font-medium text-gray-700 block mb-2">启用亲和性呼叫</label>
                <p class="text-xs text-gray-500">优化呼叫路由，提高连接质量</p>
              </div>
              <ToggleSwitch
                v-model="dialConfig.caller.affinity"
                class="ml-4"
                @update:model-value="validateDial('caller.affinity')"
              />
            </div>
            <small v-if="dialErrors.caller?.affinity" class="p-error text-xs">
              {{ dialErrors.caller?.affinity }}
            </small>

            <!-- 隐藏号码 -->
            <div class="flex items-center justify-between">
              <div class="flex-1">
                <label class="text-sm font-medium text-gray-700 block mb-2">隐藏号码</label>
                <p class="text-xs text-gray-500">在通话记录中隐藏主叫号码</p>
              </div>
              <ToggleSwitch
                v-model="privacyConfig.hideNumber"
                class="ml-4"
                @update:model-value="validatePrivacy('hideNumber')"
              />
            </div>
            <small v-if="privacyErrors.hideNumber" class="p-error text-xs">
              {{ privacyErrors.hideNumber }}
            </small>
          </div>
        </template>
        <template #footer>
          <div class="flex justify-end">
            <Button
              label="保存设置"
              :loading="savingStates.callSettings"
              icon="pi pi-save"
              size="small"
              @click="saveCallSettings"
            />
          </div>
        </template>
      </Card>

      <!-- 云端服务 -->
      <Card>
        <template #title>
          <div class="flex items-center gap-2">
            <i class="pi pi-cloud text-green-500"></i>
            <span>云端服务</span>
          </div>
        </template>
        <template #content>
          <div class="space-y-4">
            <!-- 服务器地址 -->
            <div>
              <label class="text-sm font-medium text-gray-700 block mb-2">服务器地址</label>
              <InputText
                v-model="cloudConfig.addr"
                class="w-full"
                placeholder="https://api.example.com"
                @update:model-value="validateCloud('addr')"
              />
              <small v-if="cloudErrors.addr" class="p-error text-xs">
                {{ cloudErrors.addr }}
              </small>
            </div>

            <!-- 应用ID和密钥 -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="text-sm font-medium text-gray-700 block mb-2">应用ID</label>
                <InputText
                  v-model="cloudConfig.appid"
                  class="w-full"
                  placeholder="your-app-id"
                  @update:model-value="validateCloud('appid')"
                />
                <small v-if="cloudErrors.appid" class="p-error text-xs">
                  {{ cloudErrors.appid }}
                </small>
              </div>
              <div>
                <label class="text-sm font-medium text-gray-700 block mb-2">密钥</label>
                <Password
                  v-model="cloudConfig.secret"
                  class="w-full"
                  toggle-mask
                  placeholder="输入密钥"
                  @update:model-value="validateCloud('secret')"
                />
                <small v-if="cloudErrors.secret" class="p-error text-xs">
                  {{ cloudErrors.secret }}
                </small>
              </div>
            </div>

            <!-- 前置检查选项 -->
            <div class="border-t pt-4">
              <h4 class="text-sm font-medium text-gray-700 mb-3">前置检查</h4>
              <div class="space-y-3">
                <div class="flex items-center justify-between">
                  <div class="flex-1">
                    <label class="text-sm text-gray-700">启用黑名单检查</label>
                    <p class="text-xs text-gray-500">呼叫前检查号码黑名单</p>
                  </div>
                  <ToggleSwitch
                    v-model="cloudConfig.lifecycle.precall.blacklist"
                    @update:model-value="validateCloud('lifecycle.precall.blacklist')"
                  />
                </div>
                <small v-if="cloudErrors.lifecycle?.precall?.blacklist" class="p-error text-xs">
                  {{ cloudErrors.lifecycle?.precall?.blacklist }}
                </small>

                <div class="flex items-center justify-between">
                  <div class="flex-1">
                    <label class="text-sm text-gray-700">启用闪信通知</label>
                    <p class="text-xs text-gray-500">呼叫前发送闪信提醒</p>
                  </div>
                  <ToggleSwitch
                    v-model="cloudConfig.lifecycle.precall.flashCard"
                    @update:model-value="validateCloud('lifecycle.precall.flashCard')"
                  />
                </div>
                <small v-if="cloudErrors.lifecycle?.precall?.flashCard" class="p-error text-xs">
                  {{ cloudErrors.lifecycle?.precall?.flashCard }}
                </small>
              </div>
            </div>
          </div>
        </template>
        <template #footer>
          <div class="flex justify-end">
            <Button
              label="保存"
              :loading="savingStates.cloud"
              icon="pi pi-save"
              size="small"
              @click="saveCloud"
            />
          </div>
        </template>
      </Card>
    </div>

    <!-- 右栏：ICE服务器设置 -->
    <div>
      <Card>
        <template #title>
          <div class="flex items-center gap-2">
            <i class="pi pi-server text-orange-500"></i>
            <span>ICE服务器设置</span>
          </div>
        </template>
        <template #content>
          <div class="space-y-4">
            <div>
              <label class="text-sm font-medium text-gray-700 block mb-2">JSON格式</label>
              <p class="text-xs text-gray-500 mb-2">
                '[{"urls":"stun:example.com:3478","username":"user1","credential":"pass1"}]'
              </p>
              <Textarea
                v-model="editingIceConfig"
                class="w-full font-mono text-sm"
                rows="10"
                auto-resize
                placeholder='[
  {
    "urls": "stun:example.com:3478",
    "username": "user1",
    "credential": "pass1"
  }
]'
                @update:model-value="handleIceConfig(false)"
              />
              <small
                v-if="iceErrors && Object.keys(iceErrors).length > 0"
                class="p-error text-xs block mt-2"
              >
                {{ Object.values(iceErrors)[0] }}
              </small>
            </div>
          </div>
        </template>
        <template #footer>
          <div class="flex justify-end">
            <Button
              label="保存"
              icon="pi pi-save"
              size="small"
              :loading="savingStates.ice"
              @click="handleIceConfig(true)"
            />
          </div>
        </template>
      </Card>
    </div>
  </div>
</template>
