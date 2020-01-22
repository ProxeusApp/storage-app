<template>
  <b-modal :id="'shareFileModal' + _uid"
           modal-class="modal-share-file"
           :title="$t('filegridview.share.others', 'Share File')"
           :lazy="true"
           :visible="modal"
           size="lg"
           :ok-disabled="contactsToShare.length === 0"
           :ok-title="$t('filegridview.share.button', 'Share')"
           :cancel-title="$t('generic.button.cancel', 'Cancel')"
           :busy="sharing"
           :no-close-on-backdrop="true"
           @ok="share"
           @hidden="$emit('modalClosed')">

    <div v-if="sharing" class="spinner-wrapper">
      <spinner background="transparent" :margin="80"></spinner>
    </div>
    <template v-else>
      <div class="form-group">
        <multiselect v-model="contactsToShare"
                     :options="addressesToShareFile"
                     :multiple="true"
                     track-by="address"
                     label="name"
                     :hide-selected="true"
                     :closeOnSelect="true"
                     :placeholder="$t('filegridview.share.selectContacts', 'Select contacts')"
                     :taggable="true"
                     tagPosition="bottom"
                     :tag-placeholder="$t('filegridview.share.account', 'Add ethereum account address')"
                     @tag="tag"/>

        <label class="text-muted">
          <small>{{ $t('filegridview.share.contacts', 'You can select only contacts with an attached PGP key.') }}</small>
        </label>
      </div>

      <costs v-if="gasEstimate !== undefined"
             class="mt-2"
             :gasEstimate="gasEstimate"/>
    </template>
  </b-modal>
</template>

<script>
import Spinner from '@/components/Spinner'
import Multiselect from 'vue-multiselect'
import BaseModal from '@/components/Modal/BaseModal'
import Costs from '@/components/Costs'

export default {
  name: 'share-file',
  extends: BaseModal,
  props: ['modal', 'file'],
  components: {
    Spinner,
    Multiselect,
    Costs
  },
  data () {
    return {
      sharing: false,
      contactsToShare: [],
      gasEstimate: undefined
    }
  },
  computed: {
    addressesToShareFile () {
      return this.$store.getters.addressesToShareFile(this.file)
    }
  },
  watch: {
    contactsToShare () {
      if (this.contactsToShare.length > 0) {
        this.estimateGas()
      } else {
        this.gasEstimate = undefined
      }
    }
  },
  methods: {
    async estimateGas () {
      const addresses = this.contactsToShare.map(a => a.address)
      const response = await this.$store.dispatch('SHARE_FILES_ESTIMATE_GAS', { file: this.file, addresses: addresses })

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    },
    async share (evt) {
      evt.preventDefault()
      this.sharing = true
      const addresses = this.contactsToShare.map(a => a.address)
      const response = await this.$store.dispatch('SHARE_FILES', { file: this.file, addresses: addresses })

      if (response && response.status === false && response.msg) {
        switch (response.msg) {
          case 'permission denied':
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.share.permissionDenied', 'error', {
                text: {
                  filename: this.file.filename,
                  addresses: addresses
                }
              })
            break
          case 'PGP public key missing':
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.share.pgpPublicKeyMissing', 'error', {
                text: {
                  addresses: addresses
                }
              })
            break
          case 'no new addresses provided':
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.share.noAddresses', 'error'
            )
            break
          default:
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.share.error', 'error', {
                text: {
                  filename: this.file.filename,
                  addresses: addresses
                }
              })
            break
        }
      }

      this.$emit('modalClosed')
      this.reset()
    },
    tag (searchQuery, id) {
      this.contactsToShare.push({ name: searchQuery, address: searchQuery })
    },
    reset () {
      this.sharing = false
      this.contactsToShare = []
    }
  }
}
</script>

<style lang="scss" scoped>
  .spinner-wrapper {
    min-height: 160px;
  }
</style>
