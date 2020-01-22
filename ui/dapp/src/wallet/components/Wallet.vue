<template>
  <b-modal :id="'walletModal' + _uid"
           class="walletModal"
           :visible="modal"
           :no-fade="true"
           hide-footer
           size="lg"
           @show="$store.dispatch('LOAD_BALANCE')"
           @hidden="$emit('modalClosed')">
    <div slot="modal-header" class="d-flex w-100 align-items-center">
      <h5 class="modal-title font-weight-bold">
        {{ $t('account.title', 'Your account') }}
      </h5>

      <b-dropdown id="export-dropdown" text="Dropdown Button" variant="link" class="ml-auto"
                  ref="walletSettingsDropDown" @hide="hide">
        <template slot="button-content">
          <i class="mdi md-24 mdi-settings"></i>
        </template>
        <b-dropdown-item @click="exportKeystoreToFile"
                         data-tour-step="7">{{ $t('wallet.exportKeystore', 'Export Keystore') }}
        </b-dropdown-item>
      </b-dropdown>
    </div>
    <keep-alive>
      <component :is="currentViewComponent"
                 @modalClosed="$emit('modalClosed')"
                 v-if="currentViewComponent"></component>
    </keep-alive>
  </b-modal>
</template>

<script>
/*
 * Dynamic View Components
 */
import Account from './Account'
import Login from './Auth/Login'
import { mapState } from 'vuex'

export default {
  name: 'wallet',
  props: ['mid', 'modal'],
  components: {
    Account,
    Login
  },
  data () {
    return {
      balanceCheckInterval: undefined
    }
  },
  computed: {
    ...mapState({
      authenticated: state => state.wallet.authenticated,
      unlocked: state => state.wallet.unlocked,
      balance: state => state.wallet.balance,
      walletSettingsDropDown: state => state.file.walletSettingsDropDown
    }),
    currentViewComponent () {
      return this.authenticated === true || this.unlocked === true ? Account : Login
    }
  },
  methods: {
    hide (event) {
      if (this.walletSettingsDropDown) {
        event.preventDefault()
      }
    },
    exportKeystoreToFile () {
      this.$store.dispatch('EXPORT_KEYSTORE_TO_FILE')
    }
  },
  watch: {
    walletSettingsDropDown: function (dd) {
      let ref = this.$refs.walletSettingsDropDown
      if (dd === true) {
        ref.showMenu()
        ref.visible = true
      } else {
        ref.hideMenu()
        ref.visible = false
      }
    }
  }
}
</script>

<style lang="scss" scoped>
  /deep/ .modal {
    .modal-dialog {
      position: absolute;
      right: 2rem;
      top: 1.5rem;
      width: 600px;
      transition: all 200ms !important;

      &.modal-lg {
        width: 600px;
      }

      .modal-header button {
        padding-top: 0;
        padding-bottom: 0;
      }

      .modal-content {
        box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
      }

      .modal-body {
        padding: 0;
      }
    }
  }
</style>
