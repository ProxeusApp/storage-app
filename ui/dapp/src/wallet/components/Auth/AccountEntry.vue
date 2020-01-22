<template>
  <div class="address info">
    <div class="address-info-holder" @click="currentAddress = account.address">
      <input type="text"
             ref="nameInput"
             v-model="newName"
             v-show="editing"
             @keyup.enter="saveAccountName"
             @blur="saveAccountName"
             class="edit-input"/>
      <div class="address--name" v-show="!editing" :class="{editing: currentAccount === null}">
        {{ account.name}}
      </div>
      <div class="address--hash" :class="{editing: currentAccount === null}">
        {{ account.address }}
      </div>
    </div>

    <div v-if="!currentAccount" class="actions d-flex flex-row">
      <b-dropdown variant="link" no-caret>
        <template slot="button-content">
          <i class="mdi mdi-dots-vertical md-20"></i>
        </template>
        <b-dropdown-item @click.prevent="editName">{{ $t('loginfullscreen.account.rename', 'Rename Account') }}
        </b-dropdown-item>
        <b-dropdown-item
          @click.prevent="$emit('removeAccount', account)">{{ $t('loginfullscreen.account.delete', 'Delete Account') }}
        </b-dropdown-item>
      </b-dropdown>
    </div>
    <div v-else class="actions d-flex flex-row ml-auto">
      <button class="btn d-flex align-items-center" @click.prevent="$emit('removeAccount', account)">
        <small class="change-account mr-1">{{ $t('loginfullscreen.changeAccount', 'Change account') }}</small>
        <span class="md-24 mdi mdi-arrow-down-drop-circle-outline text-primary"></span>
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AccountEntry',

  props: {
    account: {
      type: Object,
      required: false
    },
    disableRemove: {
      default: false
    }
  },
  data () {
    return {
      newName: this.account.name,
      editing: false
    }
  },
  computed: {
    currentAccount () {
      return this.$store.getters.currentAccount
    },
    currentAddress: {
      get () {
        return this.state.wallet.currentAddress
      },
      set (addr) {
        this.$store.commit('SET_CURRENT_ADDRESS', addr)
      }
    }
  },
  methods: {
    saveAccountName () {
      this.editing = false
      if (this.newName !== '') {
        this.$emit('changedName', { address: this.account.address, name: this.newName })
      }
    },
    editName () {
      this.editing = true
      this.$nextTick(() => this.$refs.nameInput.select())
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../../assets/styles/variables";

  .address {
    cursor: pointer;
    background: $gray-100;
    position: relative;
    margin-bottom: 0.5rem;

    /deep/ .actions {
      position: absolute;
      right: 0.5rem;
      top: calc(50% - 19px);

      .btn {
        transition: none;
        color: $secondary;
        background-color: $gray-100;

        &:hover {
          color: $primary;
          background-color: $gray-100;
        }
      }

      .change-account {
        opacity: 0;
        transition: 0.35s;
      }
    }

    .address-info-holder {
      border-radius: $border-radius;
      padding: 0.7rem 1rem;
      width: 100%;
    }

    &:hover {
      background-color: $gray-100;

      /deep/ .actions {
        .btn {
          color: $secondary;
          background-color: $gray-100;

          .change-account {
            opacity: 1;
            color: $secondary;
          }

          &:hover {
            color: $primary;

            .change-account {
              color: $primary;
            }
          }
        }
      }
    }
  }

  .info {
    word-wrap: break-word;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .address--name {
    word-wrap: break-word;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 250px;

    &.editing {
      max-width: 250px;
    }
  }

  .address--hash {
    color: $gray-500;
    font-size: small;
    max-width: 250px;
    overflow: hidden;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;

    &.editing {
      max-width: 250px;
    }
  }

  .edit-input {
    clear: both;
    width: 250px;
  }
</style>
