<template>
  <div class="address py-1 px-2 d-flex flex-row align-items-center" :class="{'bg-info text-white': isMyself}">
    <div class="info">
      <div class="address--name" v-show="!editing" @click="toggleNameEdit">
        {{ addressName }}
      </div>
      <input type="text"
             ref="nameInput"
             v-model="addressName"
             v-show="editing"
             @keyup.enter="saveAddressName"
             @blur="saveAddressName">
      <div class="address--hash" :class="{'text-white': isMyself}">
        {{ address.address }}
      </div>
    </div>
    <div class="actions d-flex flex-row ml-auto">
      <button class="btn btn-link p-0"
              v-if="address.pgpPublicKey === ''"
              v-tooltip="{content: $t('addressbook.contacts.tooltip.noPGPFound', 'No PGP Key found for contact'), container: '.address-book'}">
        <span class="mdi mdi-alert"></span>
      </button>
      <button class="btn btn-link p-0 ml-1"
              v-if="isMyself === false"
              @click="remove"
              v-tooltip="{content: $t('addressbook.contacts.tooltip.removeContact', 'Remove contact'), container: '.address-book'}">
        <span class="mdi mdi-minus-circle-outline"></span>
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'address-book-entry',
  props: {
    address: {
      type: Object,
      required: true
    },
    myself: {
      type: Boolean,
      default: false,
      required: false
    }
  },
  data () {
    return {
      editing: false,
      addressName: this.address.name
    }
  },
  computed: {
    isMyself () {
      return this.$store.getters.isMyself(this.address)
    }
  },
  async created () {
    if (this.address.pgpPublicKey === undefined) {
      this.$store.dispatch('UPDATE_PGPKEY', { address: this.address })
    }
  },
  methods: {
    remove () {
      this.$emit('removed', this.address)
    },
    toggleNameEdit () {
      if (this.isMyself === true) {
        return
      }
      this.editing = !this.editing
      if (this.editing) {
        this.$nextTick(() => this.$refs.nameInput.focus())
      }
    },
    saveAddressName () {
      // Only trigger changedName event when the name has changed
      if (this.address.name !== this.addressName) {
        this.$emit('changedName', { address: this.address, name: this.addressName })
      }
      this.editing = false
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .address {
    margin-bottom: 7px;
    background: $gray-200;
    border-radius: $border-radius;
  }

  .info {
    word-wrap: break-word;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .address--name {
    word-wrap: break-word;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: $primary;
  }

  .address--hash {
    color: $gray-500;
    font-size: small;
    max-width: 200px;
    overflow: hidden;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
