<template>
  <b-modal :id="'removeContactWarningModal' + _uid"
           :title="$t('filebrowser.warning.title', 'Warning')"
           :lazy="true"
           :visible="modal"
           size="lg"
           @hidden="$emit('modalClosed')">
    <p>{{ $t('addressbook.modalwarning.message', 'Do you want to remove the contact {contactName}? This action can not be undone.',
      {contactName: this.contact ? this.contact.name : ''}) }}</p>
    <div slot="modal-footer" class="row flex-fill">
        <div class="col-5 d-flex align-items-center">
          <div class="form-check">
            <input class="form-check-input" type="checkbox" :id="_uid + 'checkbox'" v-model="doNotShowAgain">
            <label class="form-check-label" :for="_uid + 'checkbox'">
              <small>{{ $t('addressbook.modalwarning.doNotShowContactRemoveWarning', 'Don\'t show this again.') }}</small>
            </label>
          </div>
        </div>
        <div class="col-7 text-right">
          <button type="button" class="btn btn-secondary mr-2" @click="$emit('modalClosed')">{{ $t('addressbook.modalwarning.cancel', 'Cancel') }}</button>
          <button type="button" class="btn btn-primary" @click="$emit('removeContact')">{{ $t('addressbook.modalwarning.removeContact', 'Remove contact') }}</button>
        </div>
    </div>
  </b-modal>
</template>

<script>
import BaseModal from '@/components/Modal/BaseModal'

export default {
  name: 'remove-contact-warning',
  extends: BaseModal,
  props: ['modal', 'contact'],
  computed: {
    doNotShowAgain: {
      get () {
        return this.$store.state.doNotShowContactRemoveWarning
      },
      set (value) {
        this.$store.commit('SET_DO_NOT_SHOW_CONTACT_REMOVE_WARNING', value)
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
