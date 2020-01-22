<template>
  <b-modal :id="'signRequestModal' + _uid"
           modal-class="modal-sign-request-file"
           :title="$t('filegridview.sign.other', 'Request others to sign this file')"
           :lazy="true"
           :visible="modal"
           size="lg"
           :ok-title="$t('filegridview.send.button', 'Send')"
           :ok-disabled="contacts.length === 0"
           :cancel-title="$t('generic.button.cancel', 'Cancel')"
           :busy="sending"
           :no-close-on-backdrop="true"
           @ok="send"
           @hidden="$emit('modalClosed')">

    <div v-if="sending" class="spinner-wrapper">
      <spinner background="transparent" :margin="80"></spinner>
    </div>
    <template v-else>
      <div class="form-group">
        <multiselect v-model="contacts"
                     :options="addressesToRequestSignatureForFile"
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
  name: 'sign-request-file',
  extends: BaseModal,
  props: ['modal', 'file'],
  components: {
    Spinner,
    Multiselect,
    Costs
  },
  data () {
    return {
      sending: false,
      contacts: [],
      gasEstimate: undefined
    }
  },
  computed: {
    addressesToRequestSignatureForFile () {
      return this.$store.getters.addressesToRequestSignatureForFile(this.file)
    }
  },
  watch: {
    contacts () {
      if (this.contacts.length > 0) {
        this.estimateGas()
      } else {
        this.gasEstimate = undefined
      }
    }
  },
  methods: {
    async estimateGas () {
      const addresses = this.contacts.map(a => a.address)
      const response = await this.$store.dispatch('SEND_SIGN_REQUEST_ESTIMATE_GAS', { file: this.file, addresses: addresses })
      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    },
    async send (evt) {
      evt.preventDefault()
      this.sending = true
      const addresses = this.contacts.map(a => a.address)
      const response = await this.$store.dispatch('SEND_SIGN_REQUEST', { file: this.file.id, addresses: addresses })
      if (response && response.status === false) {
        switch (response.msg) {
          case 'insufficient funds for gas * price + value':
            this.$store.commit('SET_INSUFFICIENT_GAS_MODAL', true)
            return
          default:
            this.$showNotification(
              'fileJS.transaction_queue.sign_request.error', 'Could not create signing request(s) for {filename}', 'error', {
                title: {
                  filename: this.file.filename
                }
              })
        }
      }

      this.$emit('modalClosed')
      this.reset()
    },
    tag (searchQuery, id) {
      this.contacts.push({ name: searchQuery, address: searchQuery })
    },
    reset () {
      this.sending = false
      this.contacts = []
    }
  }
}
</script>

<style lang="scss" scoped>
  .spinner-wrapper {
    min-height: 160px;
  }
</style>
