<template>
  <div class="file--actions" v-bind="$attrs" v-on:click.stop>
    <b-dropdown id="export-dropdown" text="Dropdown Button" variant="link" no-caret>
      <template slot="button-content">
        <i class="icon mdi md-36 mdi-dots-horizontal"></i>
      </template>
      <b-dropdown-item v-if="!inFilePreview" @click="showInfo">{{ $t('filegridview.dropdown.showInfo', 'Show Info') }}</b-dropdown-item>
      <b-dropdown-item v-if="isProcess" @click="openWorkflow">{{ $t('filegridview.dropdown.openWorkflow', 'Open Workflow') }}</b-dropdown-item>
      <b-dropdown-item v-if="!isProcess" @click="download" :disabled="isFilePendingInTxQueue">
        {{ $t('filegridview.dropdown.download', 'Download') }}
      </b-dropdown-item>
      <b-dropdown-item v-if="isFileOwner" :disabled="isFilePendingInTxQueue"
                       @click="removeFileWithWarning">{{ $t('filegridview.dropdown.remove', 'Remove file') }}
      </b-dropdown-item>
      <b-dropdown-item v-else-if="hasNoSignaturesRequired" :disabled="isFilePendingInTxQueue"
                       @click="removeFileLocalWithWarning">{{ $t('filegridview.dropdown.removeLocal', 'Locally remove file') }}
      </b-dropdown-item>
      <b-dropdown-item v-if="!isProcess" class="dropdown-item d-none">{{ $t('filegridview.dropdown.changeThumb', 'Change Thumbnail') }}
      </b-dropdown-item>
      <b-dropdown-item v-clipboard:copy="file.id"
                       v-clipboard:success="copyHashSuccess">{{ $t('filegridview.dropdown.copyFileHash', 'Copy File Hash') }}
      </b-dropdown-item>
      <b-dropdown-item v-if="isFileOwner && !isProcess" :disabled="isFilePendingInTxQueue" @click="sharePrompt">{{ $t('filegridview.dropdown.share', 'Share') }}
      </b-dropdown-item>
      <b-dropdown-item v-if="isSignable" :disabled="isFilePendingInTxQueue"
                       @click="signRequestPrompt">{{ $t('filegridview.dropdown.signRequest', 'Signing') }}
      </b-dropdown-item>
      <b-dropdown-item v-if="isFileOwner && hasReadAccess && isFileShared && !isProcess" @click="unsharePrompt" :disabled="isFilePendingInTxQueue">
        {{ $t('filegridview.dropdown.unshare', 'Unshare') }}
      </b-dropdown-item>
    </b-dropdown>
  </div>
</template>

<script>
export default {
  name: 'file-actions',
  props: {
    file: {
      type: Object,
      required: true
    },
    inFilePreview: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    isFileOwner () {
      if (this.file.owner !== undefined) {
        return this.file.owner.address === this.$store.state.wallet.currentAddress
      } else {
        return false
      }
    },
    isSignable () {
      return this.file.undefinedSignersLeft > 0 && this.isFileOwner && !this.isProcess
    },
    isProcess () {
      return this.file.fileKind === 2
    },
    hasReadAccess () {
      return this.file.readAccess !== undefined && this.file.readAccess !== null && this.file.readAccess.length > 0
    },
    hasNoSignaturesRequired () {
      return this.file.signatureStatus === 1
    },
    isFileShared () {
      let sharedWith = this.$store.getters.readAccessButNoSignatureRequestList(this.file)
      return sharedWith !== null && sharedWith.length > 0
    },
    isFilePendingInTxQueue () {
      return this.$store.getters.isFilePendingInTxQueue(this.file)
    }
  },
  methods: {
    removeFileWithWarning () {
      if (this.$store.state.file.doNotShowFileRemoveWarning === true) {
        this.$emit('removeFile')
      } else {
        this.$emit('removeFilePrompt')
      }
    },
    copyHashSuccess () {
      this.$notify({
        title: this.$t('filegridview.notify.copyClipboard', 'Copied File Hash to Clipboard'),
        text: '',
        type: 'success',
        duration: 2000
      })
    },
    async download () {
      this.$emit('download')
    },
    showInfo () {
      this.$emit('showInfo')
    },
    openWorkflow () {
      this.$store.dispatch('OPEN_PROCESS', this.file.id)
    },
    sharePrompt () {
      this.$emit('sharePrompt')
    },
    signRequestPrompt () {
      this.$emit('signRequestPrompt')
    },
    unsharePrompt () {
      this.$emit('unsharePrompt')
    },
    removeFileLocalWithWarning () {
      this.$emit('removeFileLocalPrompt')
    }
  }
}
</script>
