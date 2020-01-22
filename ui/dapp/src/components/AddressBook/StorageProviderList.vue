<template>
  <div class="address-list text-left w-100 mt-2 p-3">
    <div class="address-list">
      <storage-provider-entry v-for="sp in storageProviders" :key="sp.id"
                              :sp="sp"/>
    </div>
  </div>
</template>

<script>
import StorageProviderEntry from './StorageProviderEntry'

export default {
  name: 'storage-provider-list',
  data () {
    return {

    }
  },
  computed: {
    storageProviders () {
      return this.$store.getters.activeStorageProviders
    }
  },
  components: {
    StorageProviderEntry
  }
}
</script>
