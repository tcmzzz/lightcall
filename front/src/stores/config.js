import { defineStore } from 'pinia'
import { pb } from '@/pocketbase'

export const useConfigStore = defineStore('config', {
  state: () => ({
    iceServers: []
  }),
  actions: {
    async getIceServers() {
      try {
        const record = await pb.collection('config').getFirstListItem('name="ice_servers"')
        this.iceServers = record.value
        return this.iceServers
      } catch (err) {
        console.error('Failed to fetch ice_servers config:', err)
        return []
      }
    }
  }
})
