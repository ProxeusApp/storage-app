<script>
import moment from 'moment'
import Spinner from '../Spinner'
import TinySpinner from '../TinySpinner'
import FileActions from './FileActions'
import FILE_CONSTANTS from '../../lib/FileConstants'
import FileIcon from './FileIcon'
import FilePreview from '@/components/File/FilePreview'
import ShareFile from '@/components/Modal/ShareFile'
import SignRequestFile from '@/components/Modal/SignRequestFile'
import UnshareFile from '@/components/Modal/UnshareFile'
import RemoveFileWarning from '@/components/Modal/RemoveFileWarning'
import RemoveFileLocalWarning from '@/components/Modal/RemoveFileLocalWarning'

export default {
  name: 'file-base-view',
  components: {
    Spinner,
    TinySpinner,
    FileActions,
    FileIcon,
    FilePreview,
    ShareFile,
    SignRequestFile,
    UnshareFile,
    RemoveFileWarning,
    RemoveFileLocalWarning
  },
  props: {
    file: {
      type: Object,
      required: true
    },
    index: {
      type: Number,
      required: true
    }
  },
  data () {
    return {
      fileName: this.file.filename || this.file.id,
      generatingPreview: false,
      filePreviewModal: false,
      shareFileModal: false,
      signRequestModal: false,
      unshareFileModal: false,
      removeFileWarningModal: false,
      removeFileLocalWarningModal: false,
      FILE_CONSTANTS: FILE_CONSTANTS
    }
  },
  computed: {
    fileNames () {
      return this.$store.state.file.fileNames
    },
    addresses () {
      return this.$store.state.address.addresses
    },
    downloading () {
      return this.$store.state.file.downloading
    },
    isFileRemoving () {
      return this.$store.getters.isFileRemoving(this.file)
    },
    getFileSignStatus () {
      return this.$store.getters.fileSignStatus(this.file)
    },
    getSignersList () {
      return this.$store.getters.signersList(this.file)
    },
    getMissingSignersInfo () {
      return this.$store.getters.missingSignersInfo(this.file)
    },
    activeCategory () {
      return this.$store.state.file.activeCategory
    },
    fileExpiryFromNow () {
      return moment(this.file.expiry).fromNow()
    },
    fileGracePeriodEndFromNow () {
      return moment(this.file.expiry).add(this.file.graceSeconds, 'seconds').fromNow()
    },
    isFileSyncing () {
      return this.file.filename === this.file.id && !this.isFileExpired
    },
    isFileExpired () {
      return this.file.expired
    }
  },
  methods: {
    showPreview () {
      // Only show preview modal of no other modal is open (PCO-1341)
      if (!this.isFileSyncing && !this.removeFileWarningModal && !this.shareFileModal && !this.unshareFileModal && !this.signRequestModal) {
        this.filePreviewModal = true
      }
    },
    async removeFileLocal () {
      this.removeFileLocalWarningModal = false
      await this.$store.dispatch('REMOVE_FILE_LOCAL', this.file)
    },
    sharePrompt () {
      this.shareFileModal = true
    },
    unsharePrompt () {
      this.unshareFileModal = true
    },
    toggleFileNameEdit () {
      this.editing = !this.editing
      this.$nextTick(() => this.$refs.fileNameInput.focus())
    },
    saveFilename () {
      this.$store.commit('SET_FILENAME', { file: this.file, fileName: this.fileName })
      this.editing = false
    },
    async download () {
      this.$store.dispatch('DOWNLOAD_FILE', { file: this.file })
    },
    signRequestPrompt () {
      this.signRequestModal = true
    },
    removeFilePrompt () {
      this.removeFileWarningModal = true
    },
    async removeFile () {
      this.removeFileWarningModal = false
      if (await this.$store.dispatch('REMOVE_FILE', this.file) === false) {
        this.$showNotification(
          'fileJS.transaction_queue.remove_file.error', '', 'error', {
            title: {
              filename: this.file.filename
            }
          })
      }
    },
    removeFileLocalPrompt () {
      this.removeFileLocalWarningModal = true
    }
  }
}
</script>
