<template>
  <div class="address py-1 d-flex flex-row align-items-center" @click="showStorageProviderDetail">
    <div class="info pl-2 mr-auto">
      <div class="address--name text-truncate">
        {{ sp.name }}
      </div>
      <div class="address--address text-truncate">
        {{ sp.address }}
      </div>
    </div>
    <div class="actions pr-2 ml-auto d-flex flex-row align-items-center">
      <button class="btn btn-link px-0" @click.stop="setAsDefault"
              v-tooltip="{content: $t('addressbook.storageProvider.tooltip.selectAsDefault', 'Select as default'), container: '.address-book'}">
        <i class="mdi mdi-star" v-if="sp.address === defaultSPAddress"></i>
        <i class="mdi mdi-star-outline" v-else></i>
      </button>
      <button class="btn btn-link px-0">
        <i class="mdi mdi-chevron-right"></i>
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'storage-provider-entry',
  props: {
    sp: {
      type: Object,
      required: true
    }
  },
  computed: {
    defaultSPAddress () {
      return (this.$store.getters.defaultSP()) ? this.$store.getters.defaultSP().address : ''
    }
  },
  methods: {
    setAsDefault () {
      this.$store.dispatch('SET_DEFAULT_SP_ADDRESS', this.sp.address)
    },
    showStorageProviderDetail () {
      this.$store.commit('SET_ACTIVE_STORAGE_PROVIDER_DETAIL', { storageProvider: this.sp })
      this.$store.commit('SET_STORAGE_PROVIDER_VIEW', { view: 'detail' })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .actions .mdi {
    font-size: 1.3rem;
  }

  .address {
    margin-bottom: 7px;
    background: $info;
    border-radius: $border-radius;
    cursor: pointer;
  }

  .address--name {
    max-width: 270px;
    color: $primary;
  }

  .address--address {
    max-width: 270px;
    color: $light;
    font-size: small;
  }
</style>
