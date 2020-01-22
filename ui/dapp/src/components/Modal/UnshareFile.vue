<template>
  <b-modal :id="'unshareModal' + _uid"
           modal-class="modal-unshare-file"
           :title="$t('filegridview.unshare.others', 'Unshare wih someone previously shared')"
           :lazy="true"
           :visible="modal"
           size="lg"
           :ok-disabled="contactsToUnshare.length === 0"
           :ok-title="$t('filegridview.dropdown.unshare', 'Unshare')"
           :cancel-title="$t('generic.button.cancel', 'Cancel')"
           :busy="unsharing"
           :no-close-on-backdrop="true"
           @ok="unshare"
           @hidden="$emit('modalClosed')">

    <div v-if="unsharing" class="spinner-wrapper">
      <spinner background="transparent" :margin="80"></spinner>
    </div>
    <template v-else>
      <div class="form-group">
        <multiselect v-model="contactsToUnshare"
                     :options="contactsFileSharedWith"
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
  name: 'unshare-file',
  extends: BaseModal,
  props: ['modal', 'file'],
  components: {
    Spinner,
    Multiselect,
    Costs
  },
  data () {
    return {
      unsharing: false,
      contactsToUnshare: [],
      gasEstimate: undefined
    }
  },
  computed: {
    contactsFileSharedWith () {
      return this.$store.getters.readAccessButNoSignatureRequestList(this.file)
    }
  },
  watch: {
    contactsToUnshare () {
      if (this.contactsToUnshare.length > 0) {
        this.estimateGas()
      } else {
        this.gasEstimate = undefined
      }
    }
  },
  methods: {
    async estimateGas () {
      const addresses = this.contactsToUnshare.map(a => a.address)
      const response = await this.$store.dispatch('UNSHARE_FILE_ESTIMATE_GAS', { file: this.file, addresses: addresses })

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    },
    async unshare (evt) {
      evt.preventDefault()
      this.unsharing = true
      const addresses = this.contactsToUnshare.map(a => a.address)
      const response = await this.$store.dispatch('UNSHARE_FILE', { file: this.file, addresses: addresses })

      if (response && response.status === false && response.msg) {
        switch (response.msg) {
          case 'permission denied':
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.unshare.permissionDenied', 'error', {
                text: {
                  filename: this.file.filename,
                  addresses: addresses
                }
              })
            break
          case 'PGP public key missing':
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.unshare.pgpPublicKeyMissing', 'error', {
                text: {
                  addresses: addresses
                }
              })
            break
          default:
            this.$showNotification(
              'general.notify.titleError', 'fileJS.transaction_queue.unshare.error', 'error', {
                text: {
                  filename: this.file.filename
                }
              })
            break
        }
      }

      this.$emit('modalClosed')
      this.reset()
    },
    tag (searchQuery, id) {
      this.contactsToUnshare.push({ name: searchQuery, address: searchQuery })
    },
    reset () {
      this.unsharing = false
      this.contactsToUnshare = []
    }
  }
}
</script>

<style lang="scss" scoped>
  .spinner-wrapper {
    min-height: 160px;
  }
</style>
