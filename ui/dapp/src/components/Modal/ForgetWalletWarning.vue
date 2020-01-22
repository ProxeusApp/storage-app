<template>
  <b-modal :id="'forgetWalletWarningModal' + _uid"
           :title="$t('filebrowser.warning.title', 'Warning')"
           :lazy="true"
           :visible="modal"
           @hidden="closeModal">
    <spinner v-if="removing || exporting" background="transparent" :margin="70" style="position: relative;"></spinner>

    <div v-if="!removing && !exporting">
      <p>{{ $t('forgetwallet.modal.warning', 'If you have not exported your keystore, you will lose your wallet and we will not be able to restore it.') }}</p>
      <p>{{ $t('forgetwallet.modal.password', 'Please enter your password to export your keystore or forget your wallet.') }}</p>
      <div class="form-group">
        <input type="password" class="mt-1 form-control" v-model="password" name="password"
               ref="passwordInput" :placeholder="$t('loginfullscreen.password', 'Password')">
      </div>
      <div class="auth-error" v-if="authError">
        <span class="text-danger">{{ authError }}</span>
      </div>
    </div>
    <div slot="modal-footer" class="flex-wrap w-100">
      <div class="row flex-fill">
        <div class="col-5 col-lg-auto pb-1 pb-lg-0">
          <button type="button" :disabled="removing || exporting || password < 1" class="btn btn-primary"
                  @click="exportKeystoreByAddress">{{ $t('wallet.exportKeystore', 'Export Keystore') }}
          </button>
        </div>
        <div class="col-7 col-lg-auto ml-lg-auto text-right">
          <button type="button" :disabled="removing || exporting" class="btn btn-secondary mr-1"
                  @click="closeModal">{{ $t('generic.button.cancel', 'Cancel') }}
          </button>
          <button type="button" :disabled="removing || exporting || password < 1" class="btn btn-danger"
                  @click="removeAccount">{{ $t('filebrowser.buttons.forgetWallet', 'Forget Wallet') }}
          </button>
        </div>
      </div>
    </div>
  </b-modal>
</template>

<script>
import BaseModal from '@/components/Modal/BaseModal'
import Spinner from '@/components/Spinner'

export default {
  name: 'forget-wallet-warning',
  extends: BaseModal,
  components: {
    Spinner
  },
  props: [
    'modal',
    'accountToRemove'
  ],
  data () {
    return {
      password: '',
      authError: '',
      removing: false,
      exporting: false
    }
  },
  computed: {
    authenticated () {
      return this.$store.state.wallet.authenticated
    }
  },
  watch: {
    modal: function (modal) {
      setTimeout(() => { // nextTick does not work here
        if (modal === true) {
          this.$refs.passwordInput.focus()
        }
      })
    }
  },
  methods: {
    async exportKeystoreByAddress () {
      this.exporting = true
      let res = await this.$store.dispatch('EXPORT_KEYSTORE_BY_ADDRESS', {
        ethAddress: this.accountToRemove,
        password: this.password
      })

      if (res.status === false && res.msg === 'permission denied') {
        this.exporting = false
        this.authError = this.$t('loginfullscreen.invalidPassword', 'You have entered an invalid password.')
        setTimeout(() => this.$refs.passwordInput.focus())
        return
      }

      if (res.status === false) {
        // TODO Send an appropriate error message according to error code
        this.authError = res.msg
        return
      }
      this.exporting = false
      this.password = ''
      this.authError = ''
    },
    closeModal () {
      this.authError = ''
      this.password = ''
      this.removing = false
      this.exporting = false
      this.$emit('modalClosed')
    },
    async removeAccount () {
      this.removing = true
      let res = await this.$store.dispatch('REMOVE_ACCOUNT', {
        ethAddress: this.accountToRemove,
        password: this.password
      })

      if (res.status === false) {
        if (res.msg === 'permission denied') {
          this.removing = false
          this.authError = this.$t('loginfullscreen.invalidPassword', 'You have entered an invalid password.')
          setTimeout(() => this.$refs.passwordInput.focus())
          return
        }
        this.authError = res.msg
        return
      }

      this.closeModal()
    }
  }
}
</script>
