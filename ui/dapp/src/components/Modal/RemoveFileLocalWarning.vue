<template>
  <b-modal :id="'removeFileLocalWarningModal' + _uid"
           :title="$t('filebrowser.warning.title', 'Warning')"
           :lazy="true"
           :visible="modal"
           :ok-title="$t('filebrowser.buttons.removeFileLocal', 'Locally remove file')"
           :cancel-title="$t('generic.button.cancel', 'Cancel')"
           :no-close-on-backdrop="true"
           @ok="$emit('removeFileLocal')"
           @hidden="$emit('modalClosed')">

    <p>{{ $t('filebrowser.warning.removeFileLocal', 'Do you want to locally remove the file {filename}? This action can not be undone.', {filename: this.file.filename}) }}</p>
  </b-modal>
</template>

<script>
import BaseModal from '@/components/Modal/BaseModal'

export default {
  name: 'remove-file-local-warning',
  extends: BaseModal,
  props: ['modal', 'file']
}
</script>

<style scoped>
  /deep/ .modal-body {
    word-wrap: break-word;
  }
</style>
