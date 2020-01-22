<template>
  <div class="file-upload-container">
    <b-modal :id="'processModal' + _uid"
             :title="$t('filebrowser.fileupload.share_title', 'Share Process')"
             :lazy="true"
             :visible="modal"
             size="lg"
             :ok-title="$t('filebrowser.fileupload.share_confirm', 'Confirm')"
             :cancel-title="$t('generic.button.cancel', 'Cancel')"
             :busy="uploading"
             @ok="upload"
             @hidden="close">
      <div class="row" style="min-height:200px;" v-if="uploading">
        <spinner background="transparent"></spinner>
      </div>
      <div class="row" v-if="!uploading">
        <div class="col-md-8">
          <div class="form-group">
            <label>{{ $t('filebrowser.fileupload.select_storage_provider.label', 'Select storage provider') }}
            </label>
            <multiselect v-model="sProviderHandler"
                         :options="storageProviders"
                         :multiple="false"
                         track-by="name"
                         label="name"
                         :searchable="false"
                         :closeOnSelect="true"
                         :show-labels="false"
                         :placeholder="$t('filebrowser.fileupload.select_storage_provider.placeholder', 'Select storage provider')"/>
            <small class="text-muted">{{ $t('filebrowser.fileupload.help_text_sp', 'Help Text') }}</small>
          </div>
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script>
import Spinner from '@/components/Spinner'
import Multiselect from 'vue-multiselect'
import BaseModal from '@/components/Modal/BaseModal'

export default {
  name: 'process-component',
  extends: BaseModal,
  props: ['modal'],
  components: {
    Multiselect,
    Spinner
  },
  data () {
    return {
      newFileHash: undefined,
      uploading: false,
      sProvider: undefined
    }
  },
  computed: {
    // processInfo () {
    //   return this.$store.getters.processInfo
    // },
    insufficientEtherBalanceText () {
      return this.$t('filebrowser.fileupload.insufficientEtherBalance',
        'Your current balance of {ethBalance} ETH is insufficient for this transaction.  Please top up your Proxeus Wallet. Get more <a href="http://faucet.ropsten.be:3001/">Ether tokens</a>',
        { ethBalance: this.ethBalance })
    },
    allowance () {
      return this.$store.state.wallet.allowance
    },
    ethBalance () {
      return this.$store.state.wallet.ethBalance
    },
    addressesWithPGPKey () {
      return this.$store.getters.addressesWithPGPKey
    },
    storageProviders () {
      return this.$store.getters.activeStorageProviders
    },
    sProviderHandler: {
      get () {
        return this.sProvider || this.$store.getters.defaultSPOrFirst()
      },
      set (sProvider) {
        this.sProvider = sProvider
      }
    },
    insufficientGasEstimationModal: {
      get () {
        return this.$store.state.notification.insufficientGasEstimationModal
      },
      set (showModal) {
        this.$store.commit('SET_INSUFFICIENT_GAS_MODAL', showModal)
      }
    },
    insufficientXesAllowanceModal: {
      get () {
        return this.$store.state.notification.insufficientXesAllowanceModal
      },
      set (showModal) {
        this.$store.commit('SET_INSUFFICIENT_XES_ALLOWANCE_MODAL', showModal)
      }
    }
  },
  methods: {
    /**
     * Start the registration/upload process
     */
    async upload (evt) {
      evt.preventDefault()
      this.uploading = true
      if (!this.sProviderHandler) {
        this.$showNotification('filebrowser.notify.error', 'filebrowser.notify.storageProviderNotFound', 'error')
        this.uploading = false
        return
      }

      let formData = new FormData()
      formData.append('link', window.shareprocesslink)

      console.log(formData)
      if (this.sProviderHandler) {
        formData.append('providerAddress', this.sProviderHandler.address)
      }
      formData.append('duration', 14)

      const response = await this.$store.dispatch('SHARE_PROCESS', formData)

      if (response.status === false && response.msg) {
        this.close()
        switch (response.msg) {
          case 'file already exists':
            this.$showNotification('general.notification.error.title', 'filebrowser.notify.fileRegistered', 'error')
            this.close()
            return
          case 'gas required exceeds allowance or always failing transaction':
            this.$emit('insufficientXesAllowance', 0.3)
            this.close()
            return
          case 'insufficient funds for gas * price + value':
            this.insufficientGasEstimationModal = true
            this.close()
            return
          case 'PGP public key missing':
            this.$showNotification('general.notification.error.title', 'filebrowser.notify.noPgpAttached', 'error')
            this.close()
            return
          case 'file exceeds file size limit':
            this.$showNotification('general.notification.error.title', 'filebrowser.notify.fileExceedsFileSizeLimit',
              'error')
            this.close()
            return
          default:
            this.$showNotification('general.notification.error.title', 'fileJS.upload.couldNotUpload', 'error')
            this.$emit('modalClosed')
            return
        }
      }

      this.$emit('uploaded')
      this.close()
    },
    close () {
      this.$emit('modalClosed')
      this.reset()
    },
    reset () {
      this.recipient = undefined
      this.uploading = false
    }
  }
}
</script>

<style lang="scss" scoped>
  @media (min-width: 576px) {
    /deep/ .modal-dialog {
      max-width: 700px;
    }
  }
</style>
