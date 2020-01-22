<template>
  <b-modal :id="'removeFileWarningModal' + _uid"
           :title="$t('filebrowser.warning.title', 'Warning')"
           :lazy="true"
           :visible="modal"
           :no-close-on-backdrop="true"
           @hidden="$emit('modalClosed')">

    <p>{{ $t('filebrowser.warning.removeFile', 'Do you want to remove the file {filename}? This action can not be undone.', {filename: this.file.filename}) }}</p>

    <costs v-if="gasEstimate !== undefined"
           class="mt-3"
           :gasEstimate="gasEstimate"/>

    <div slot="modal-footer" class="row flex-fill">
      <div class="col-5 d-flex align-items-center pr-0">
        <div class="form-check">
          <input class="form-check-input" type="checkbox" :id="_uid + 'checkbox'" v-model="doNotShowAgain">
          <label class="form-check-label" :for="_uid + 'checkbox'">
            <small>{{ $t('filebrowser.warning.doNotShowFileRemoveWarning', 'Don\'t show this again.') }}</small>
          </label>
        </div>
      </div>
      <div class="col-7 text-right">
        <button type="button" class="btn btn-secondary"
                @click="$emit('modalClosed')">{{ $t('generic.button.cancel', 'Cancel') }}
        </button>
        <button type="button" class="btn btn-primary ml-2"
                @click="$emit('removeFile')">{{ $t('filebrowser.buttons.removeFile', 'Remove file') }}
        </button>
      </div>
    </div>
  </b-modal>
</template>

<script>
import BaseModal from '@/components/Modal/BaseModal'
import Costs from '@/components/Costs'

export default {
  name: 'remove-file-warning',
  extends: BaseModal,
  components: {
    Costs
  },
  props: ['modal', 'file'],
  data () {
    return {
      gasEstimate: undefined
    }
  },
  computed: {
    doNotShowAgain: {
      get () {
        return this.$store.state.doNotShowFileRemoveWarning
      },
      set (value) {
        this.$store.commit('SET_DO_NOT_SHOW_FILE_REMOVE_WARNING', value)
      }
    }
  },
  watch: {
    modal () {
      if (this.modal) {
        this.estimateGas()
      }
    }
  },
  methods: {
    async estimateGas () {
      const response = await this.$store.dispatch('REMOVE_FILE_ESTIMATE_GAS', this.file)

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    }
  }
}
</script>

<style scoped>
  /deep/ .modal-body {
    word-wrap: break-word;
  }
</style>
