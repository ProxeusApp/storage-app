<template>
  <notification-entry :id="'notification-'+ notification.id"
                      :notification="notification"
                      :key="notification.id">
    <b-popover :target="'notification-'+ notification.id +'-btn-sign'"
               ref="signEstimateGasPopover"
               placement="left"
               triggers="hover"
               :container="'notification-'+ notification.id"
               :disabled="!notification.pending || signingInProgress"
               @show="onPopoverShow">
      <costs v-if="gasEstimate !== undefined"
             :gasEstimate="gasEstimate"/>
      <div v-else class="card">
        <ul class="list-group list-group-flush">
          <li class="list-group-item">
            <h5 class="mb-0">{{ $t('filebrowser.notifications.estimateGas', 'Estimate gas…') }}</h5>
          </li>
        </ul>
      </div>
    </b-popover>
    <button type="button" class="btn btn-light btn-sm px-1 py-1"
            :id="'notification-'+ notification.id +'-btn-sign'"
            :disabled="!notification.pending || isFilePendingInTxQueue"
            @click="sign">
      {{ signActionText }}
      <div v-if="signingInProgress" class="spinner ml-1 d-block">
        <div class="tinyspinner"></div>
      </div>
    </button>
    <button type="button" class="btn btn-light btn-sm px-1 py-1"
            :disabled="isFilePendingInTxQueue || notification.data.fileRemoved === true"
            @click="download">
      {{ $t('filebrowser.fileupload.download','Download') }}
    </button>
  </notification-entry>
</template>

<script>
import NotificationEntry from '@/components/File/NotificationEntry'
import Costs from '@/components/Costs'

export default {
  name: 'signing-entry',
  props: ['notification'],
  data () {
    return {
      gasEstimate: undefined
    }
  },
  components: {
    NotificationEntry,
    Costs
  },
  computed: {
    isFilePendingInTxQueue () {
      return this.$store.getters.isFileHashPendingInTxQueue(this.notification.data.hash)
    },
    signingInProgress () {
      return this.notification.actionInProgress !== undefined && this.notification.actionInProgress === 'signing'
    },
    signActionText () {
      return this.signingInProgress
        ? this.$t('filebrowser.notifications.notification_signing', 'Signing')
        : this.$t('filebrowser.fileupload.sign_files.sign', 'Sign')
    },
    virtualFile () {
      let virtualFile = { ...this.notification.data }
      virtualFile.id = virtualFile.hash
      virtualFile.filename = virtualFile.fileName
      return virtualFile
    }
  },
  methods: {
    onPopoverShow (event) {
      if (this.notification.pending && !this.signingInProgress) {
        this.estimateGas()
      } else {
        event.preventDefault()
      }
    },
    async estimateGas () {
      const response = await this.$store.dispatch('SIGN_HASH_ESTIMATE_GAS', this.notification.data.hash)

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    },
    async sign () {
      this.$refs.signEstimateGasPopover.$emit('close')

      this.$store.commit('ADD_NOTIFICATION_ACTION_IN_PROGRESS',
        { 'notification': this.notification, 'action': 'signing' })

      const response = await this.$store.dispatch('SIGN_HASH', this.notification.data.hash)

      if (response && response.status === false) {
        switch (response.msg) {
          case 'insufficient funds for gas * price + value':
            this.$store.commit('SET_INSUFFICIENT_GAS_MODAL', true)
            return
          default:
            this.$showNotification(
              'general.notify.titleError',
              'fileJS.transaction_queue.sign_file.error', 'error', {
                text: {
                  filename: this.notification.data.fileName
                }
              })
        }
      }
    },
    download () {
      this.$store.dispatch('DOWNLOAD_FILE', { file: this.virtualFile })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables.scss";

  /deep/ .popover {
    width: 350px;
    max-width: none;
  }

  /deep/ .popover-body {
    padding: 0;
  }

  .cursor-default {
    cursor: default;
  }

  .file-name,
  .address-name {
    font-size: 0.8rem;
  }

  .trim {
    word-wrap: break-word;
    overflow: hidden;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .spinner {
    min-width: 1rem;
    height: 1rem;
    display: inline-block;
    float: right;
    padding-top: 4px;
  }

  .tinyspinner {
    display: inline-block;
    width: 1rem;
    height: 1rem;
    border: 0.15rem solid $gray-500;
    border-bottom: 0.15rem solid rgba(0, 0, 0, 0);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    z-index: 9999;
  }

  .tinyspinner--hidden {
    display: none;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  @-webkit-keyframes rotating {
    from {
      -webkit-transform: rotate(0deg);
    }

    to {
      -webkit-transform: rotate(360deg);
    }
  }
</style>
