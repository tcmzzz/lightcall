<script setup>
import { useRoute } from 'vue-router'
const route = useRoute()
defineProps({ menuitems: { type: Object, required: true } })
</script>

<template>
  <div class="h-full flex flex-col">
    <Tabs :value="route.path">
      <TabList>
        <Tab v-for="tab in menuitems" :key="tab.label" :value="tab.route">
          <router-link v-if="tab.route" v-slot="{ href, navigate }" :to="tab.route" custom>
            <a v-ripple :href="href" class="flex items-center gap-2 text-inherit" @click="navigate">
              <i :class="tab.icon" />
              <span>{{ tab.label }}</span>
            </a>
          </router-link>
        </Tab>
      </TabList>
    </Tabs>
    <div class="flex-1 flex flex-col overflow-hidden mx-10 mt-4">
      <RouterView />
    </div>
  </div>
</template>
