<template>
  <!-- dragover.prevent is needed below to define a drop zone -->
  <div class="file-browser" @dragover.prevent @dragenter.prevent="drag" @dragleave.prevent="drag"
       @drop.prevent="fileDropHandler">
    <file-drag-overlay :is-file-dragging="isFileDragging"/>
    <wallet-component mid="wallet-component" :modal="walletModal"
                      @modalClosed="walletModal = false"></wallet-component>
    <process-component mid="process-component" :modal="processModal"
                       @modalClosed="processModal = false"></process-component>
    <top-nav :showAddressBook="showAddressBook"
             @exportedInfo="exportInfo"
             @onLock="logout"
             @toggledShowAddressBook="showAddressBook = !showAddressBook"
             @notificationModalOpened="notificationModal = true"
             @walletModalOpened="walletModal = true"
             @processModalOpened="processModal = true"></top-nav>
    <div class="container-fluid">
      <div class="row">
        <secondary-nav @dropped="dropped" @toggleSyncRow="syncVisible = !syncVisible"></secondary-nav>
      </div>
      <backdrop :visible="syncVisible === true" @clicked="syncVisible = false" :zindex="1000"></backdrop>
      <sync-row :syncVisible="syncVisible"></sync-row>
      <div class="row main-content">
        <category-sidebar @changeActiveCategory="changeActiveCategory"></category-sidebar>
        <div class="main">
          <h2 v-if="filesLoadingError === false">{{ categoryTitle }}</h2>
          <file-list :files="categorizedFiles"
                     :filesLoading="filesLoading" v-if="filesLoadingError === false"></file-list>
          <div class="w-100 text-center mt-3" v-else>
            <h2 class="text-muted">{{ $t('filebrowser.fileList.loadingError', 'Could not load files') }}</h2>
            <button class="btn btn-primary"
                    @click="loadFiles">{{ $t('filebrowser.fileList.retryLoadingFiles', 'Retry') }}
            </button>
          </div>
        </div>
      </div>
      <backdrop :visible="showAddressBook === true" @clicked="showAddressBook = false"></backdrop>
      <div class="col-address-book" :class="{active: showAddressBook}">
        <address-book @closeAddressBook="showAddressBook = false"></address-book>
      </div>
    </div>
    <b-modal v-model="insufficientGasEstimationModal" ok-only
             :title="$t('filebrowser.fileupload.insufficientEtherBalanceTitle', 'Insufficient Ether Balance')"
             @hidden="insufficientGasEstimationModal = false">
      <p v-html="insufficientEtherBalanceText"></p>
    </b-modal>
    <b-modal v-model="insufficientXesAllowanceModal" ok-only
             :title="$t('filebrowser.fileupload.insufficientXesAllowanceTitle', 'Insufficient XES Allowance')"
             @hidden="insufficientXesAllowanceModal = false">
      {{ $t('filebrowser.fileupload.insufficientXesAllowance', 'This transaction will cost you {xesTransactionCost} XES. Currently you have a PSPP allowance of {allowance} XES which is insufficient. Please increase the PSPP allowance in your wallet', {xesTransactionCost: xesTransactionCost,
      allowance: allowance}) }}
    </b-modal>
    <b-modal v-model="insufficientXesModal" ok-only
             :title="$t('filebrowser.fileupload.insufficientXesTitle', 'Insufficient XES')"
             @hidden="insufficientXesModal = false">
      {{ $t('filebrowser.fileupload.insufficientXes', 'This transaction will cost you {xesTransactionCost} XES. Currently you have a balance of {xesBalance} XES which is insufficient. Please deposit more XES to your wallet.', {xesTransactionCost: xesTransactionCost,
      xesBalance: xesBalance}) }}
    </b-modal>
    <file-upload :newFile="newFile"
                 :modal="fileUploadModal"
                 @uploaded="fileUploadModal = undefined; newFile = undefined"
                 @insufficientXesAllowance="insufficientXesAllowance"
                 @insufficientXes="insufficientXes"
                 @modalClosed="fileUploadModal = false; newFile = undefined"></file-upload>
    <notification-modal :modal="notificationModal"
                        :title="$t('filebrowser.fileupload.Notifications', 'Notifications')"
                        @modalClosed="notificationModal = false"></notification-modal>
    <remove-contact-warning :contact="contactToRemove"
                            :modal="removeContactWarningModal"
                            @modalClosed="closeRemoveContactWarningModal"
                            @removeContact="removeContact"></remove-contact-warning>
    <tour></tour>
    <upgrade-version :modal="upgradeModal" :version="version" @modalClosed="upgradeModalClose"></upgrade-version>
  </div>
</template>

<script>
import AddressBook from '@/components/AddressBook/AddressBook'
import Backdrop from '@/components/Backdrop'
import FileList from '@/components/File/FileList'
import FileUpload from '@/components/Modal/FileUpload'
import ProcessComponent from '@/components/Modal/ProcessModal'
import TopNav from '@/components/TopNav'
import SecondaryNav from '@/components/SecondaryNav'
import SyncRow from '@/components/File/SyncRow'
import CategorySidebar from '@/components/CategorySidebar'
import NotificationModal from '@/components/Modal/NotificationModal'
import UpgradeVersion from '@/components/Modal/UpgradeVersion'
import Tour from '@/components/Tour'
import RemoveContactWarning from '@/components/Modal/RemoveContactWarning'

import { mapState } from 'vuex'

import WalletComponent from '../wallet/components/Wallet'
import FileDragOverlay from '../components/File/FileDragOverlay'

export default {
  name: 'file-browser',
  components: {
    Backdrop,
    AddressBook,
    FileDragOverlay,
    FileList,
    FileUpload,
    TopNav,
    SecondaryNav,
    SyncRow,
    CategorySidebar,
    'wallet-component': WalletComponent,
    'process-component': ProcessComponent,
    NotificationModal,
    UpgradeVersion,
    Tour,
    RemoveContactWarning
  },
  data () {
    return {
      notificationModal: false,
      fileUploadModal: false,
      upgradeModal: false,

      showAddressBook: false,
      showFileUploader: true,

      signersCount: 0,
      syncVisible: false,
      loadFilesInterval: undefined,
      fileSyncInterval: undefined,
      xesTransactionCost: '',
      transactionCost: undefined,
      newFile: undefined,
      dragEventCount: 0
    }
  },
  computed: {
    files () {
      return this.$store.getters.filteredFiles
    },
    sharedWithMe () {
      return this.$store.getters.sharedWithMe(this.$store.state.wallet.currentAddress, this.searchTerm)
    },
    activeCategory () {
      return this.$store.state.file.activeCategory
    },
    categoryTitle () {
      switch (this.activeCategory) {
        case 'all-files':
          return this.$t('filebrowser.sidebar.all_files', 'All Files')
        case 'my-files':
          return this.$t('filebrowser.sidebar.my_files', 'My files')
        case 'shared-with-me':
          return this.$t('filebrowser.sidebar.shared_with_me', 'Shared with me')
        case 'signed-by-me':
          return this.$t('filebrowser.sidebar.signed_by_me', 'Signed by me')
        case 'expired-files':
          return this.$t('filebrowser.sidebar.expired_files', 'Expired files')
      }

      return ''
    },
    categorizedFiles () {
      return this.files
    },
    airdropHintOutstanding () {
      return this.$store.state.airdropHintOutstanding.find(a => a === this.currentAddress) !== undefined
    },
    filesLoading () {
      return this.$store.getters.filesLoading
    },
    isFileDragging () {
      return this.dragEventCount > 0
    },
    insufficientEtherBalanceText () {
      return this.$t('filebrowser.fileupload.insufficientEtherBalance',
        'Your current balance of {ethBalance} ETH is insufficient for this transaction.  Please top up your Proxeus Wallet. Get more <a href="http://faucet.ropsten.be:3001/">Ether tokens</a>',
        { ethBalance: this.ethBalance })
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
    },
    insufficientXesModal: {
      get () {
        return this.$store.state.notification.insufficientXesModal
      },
      set (showModal) {
        this.$store.commit('SET_INSUFFICIENT_XES_MODAL', showModal)
      }
    },
    walletModal: {
      get () {
        return this.$store.state.file.walletModal
      },
      set (showModal) {
        this.$store.commit('SET_WALLET_MODAL', showModal)
      }
    },
    processModal: {
      get () {
        var processModalState = this.$store.state.file.processModal
        if (processModalState === true) {
          this.$emit('processModalOpened')
        }
        return processModalState
      },
      set () {
        this.$store.commit('SET_PROCESS_MODAL', false)
      }
    },
    ...mapState({
      addresses: state => state.address.addresses,
      allowance: state => state.wallet.allowance,
      xesBalance: state => state.wallet.balance,
      currentAddress: state => state.wallet.currentAddress,
      ethBalance: state => state.wallet.ethBalance,
      availableLanguages: state => state.availableLanguages,
      filesQueue: state => state.file.filesQueue,
      filesLoadingError: state => state.file.filesLoadingError,
      transactionQueue: state => state.file.transactionQueue,
      version: state => state.version,
      upgradeDismissed: state => state.upgradeDismissed,
      contactToRemove: state => state.address.addressToRemove,
      removeContactWarningModal: state => state.address.removeContactWarningModal,
      logoutInProgress: state => state.wallet.logoutInProgress
    })
  },
  async mounted () {
    try {
      this.$store.dispatch('INIT_CHANNEL_HUB')
      this.$store.dispatch('LOAD_ADDRS')
      this.$store.dispatch('LOAD_STORAGE_PROVIDERS')
    } catch (e) {
      console.log(e)
    }
  },
  methods: {
    changeActiveCategory (category) {
      this.$store.commit('SET_ACTIVE_CATEGORY', category)
      this.$store.dispatch('LOAD_FILES')
    },
    loadFiles () {
      this.$store.dispatch('LOAD_FILES')
    },
    loadBalance () {
      this.$store.dispatch('LOAD_BALANCE')
    },
    productTourStopCallback () {
      this.$store.commit('SET_PRODUCT_TOUR_COMPLETED', this.currentAddress)
    },
    exportInfo () {
      this.$store.dispatch('EXPORT_INFO_TO_FILE')
    },
    importInfo (fileList) {
      this.$store.dispatch('IMPORT_INFO_FROM_FILE', { file: fileList[0] }).then(response => {
        this.$notify({
          title: this.$t('general.notify.titleSuccess', 'Info succesfully imported'),
          type: 'success'
        })
      }, () => {
        this.$notify({
          title: this.$t('general.notify.titleError', 'Error'),
          text: this.$t('general.notify.errorImportFile', 'Could not import file'),
          type: 'error'
        })
      })
    },
    async fileDropHandler (event) {
      this.dropped(event.dataTransfer.files[0])
    },
    async dropped (file) {
      // Dragging & dropping selected text causes "file" to be undefined
      if (file) {
        this.fileUploadModal = true
        this.newFile = file
      }
      this.dragEventCount = 0
    },
    drag (event) {
      if (event.type === 'dragenter') {
        this.dragEventCount++
      }
      if (event.type === 'dragleave') {
        this.dragEventCount--
      }
    },
    insufficientXesAllowance (xesTransactionCost) {
      this.xesTransactionCost = xesTransactionCost
      this.insufficientXesAllowanceModal = true
    },
    insufficientXes (xesTransactionCost) {
      this.xesTransactionCost = xesTransactionCost
      this.insufficientXesModal = true
    },
    logout () {
      this.$store.dispatch('CLOSE_CHANNEL_HUB')
      this.$store.dispatch('LOAD_ACCOUNTS_AND_SET_FIRST_ACTIVE') // reload to get updated list oder
    },
    upgradeModalClose () {
      this.upgradeModal = false
      this.$store.commit('UPGRADE_DISMISSED')
    },
    closeRemoveContactWarningModal () {
      if (!this.logoutInProgress) {
        this.showAddressBook = true
      }
      this.$store.commit('SET_REMOVE_CONTACT_WARNING_MODAL', false)
    },
    removeContact () {
      this.showAddressBook = true
      this.$store.dispatch('REMOVE_ADDRESS', this.$store.state.address.addressToRemove)
      this.$store.commit('SET_REMOVE_CONTACT_WARNING_MODAL', false)
    }
  },
  watch: {
    version: function (version) {
      if (version && version.update !== 'none' && !this.upgradeDismissed) {
        this.upgradeModal = true
      }
    }
  }
}

</script>

<style lang="scss">
  @import "../assets/styles/variables";
  @import "~vue-multiselect/dist/vue-multiselect.min.css";

  .main-content {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-areas: "sidebar main";
    padding-top: calc(55px + 66px); // topnav height + secondary-nav height
  }

  .main {
    padding: 1.5rem;
    grid-area: main;
    overflow-x: auto;
  }

  .multiselect__select {
    top: 50%;
    height: auto;

    .multiselect--active & {
      display: none;
    }
  }

  .multiselect__input {
    margin-bottom: 7px;
    padding: 0;
    line-height: calc(1rem + 1rem); // font-size + tag padding-top/bottom

    &::placeholder {
      color: $text-muted;
    }

    .multiselect--active & {
      .modal-file-upload & {
        min-width: 360px;
      }

      .modal-share-file &,
      .modal-unshare-file &,
      .modal-sign-request-file & {
        min-width: 320px;
      }
    }
  }

  .multiselect__placeholder {
    margin-bottom: 7px;
    padding-top: 0;
    color: $text-muted;
    line-height: calc(1rem + 1rem); // font-size + tag padding-top/bottom
  }

  .multiselect__tags-wrap {
    display: inline-block;
  }

  .multiselect__tag {
    margin-bottom: 0;
    padding: 0.5rem 30px 0.5rem 0.7rem;
    border-radius: $border-radius;
    background-color: $info;

    > span {
      display: inline-block;
      max-width: 200px;
      text-overflow: ellipsis;
      overflow: hidden;
      vertical-align: middle;
      font-size: 90%;
    }

    .multiselect__tag-icon {
      line-height: calc(1rem + 1rem); // font-size + tag padding-top/bottom

      &:after {
        font-size: 15px;
      }

      &:hover {
        background: darken($info, 10%);
      }
    }
  }

  .multiselect__option {
    &.multiselect__option--selected {
      background: $gray-300;

      &.multiselect__option--highlight,
      &.multiselect__option--highlight:after {
        background: #062a85;
      }
    }

    &.multiselect__option--highlight,
    &.multiselect__option--highlight:after {
      background: $primary;
    }
  }

  .multiselect__content-wrapper {
    box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
  }

  .flatpickr-input {
    line-height: 2;
    padding: 0.5rem 1rem;
  }

  .multiselect__tags,
  .flatpickr-input {
    display: block;
    font-size: 1rem;
    color: #495057;
    background-color: #ffffff;
    background-clip: padding-box;
    border: 2px solid $gray-300 !important;
    border-radius: 0.25rem;
    transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;

    &:focus {
      border-color: rgba(6, 42, 133, 0.6) !important;
    }
  }
</style>

<style lang="scss" scoped>
  @import "../assets/styles/variables";

  .col-main {
    transition: all 200ms;
  }

  .col-address-book {
    position: fixed;
    top: 3.5rem;
    right: 1rem;
    width: 400px;
    height: auto;
    transition: transform 200ms ease;
    transform: scale(0);
    transform-origin: top right;
    z-index: 1100;
    border-radius: $border-radius;

    &.active {
      transition: transform 200ms ease;
      transform: scale(1);
      transform-origin: top right;
      box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
    }
  }

  .fixed-height-scroll {
    height: calc(100vh - 62px);

    .col-main,
    .col-address-book {
      overflow-y: auto;
    }
  }

  .col-address-import-export {
    position: fixed;
    top: 4rem;
    right: 1rem;
    width: 350px;
    height: auto;
    max-height: calc(100vh - 80px);
    overflow: auto;
    transition: transform 200ms ease;
    transform: scale(0);
    transform-origin: top right;
    z-index: 1100;
    border-radius: $border-radius;

    &.active {
      top: 8rem;
      transition: transform 200ms ease;
      transform: scale(1);
      transform-origin: top right;
      box-shadow: 0 40px 100px rgba(0, 0, 0, 0.35);
    }
  }

  .file--status {
    font-size: $font-size-sm;
  }

  .global-spinner {
    background: rgba(0, 0, 0, 0.1);
    position: fixed;
    width: 100%;
    height: 100%;
    top: 0;
    left: 0;
    z-index: 10000;
  }

  .dropdown-item {
    position: relative;
    cursor: pointer;
  }

  .btn-userdata-upload {
    position: relative;
  }

  .file-userdata-upload {
    position: absolute;
    font-size: 50px;
    opacity: 0;
    right: 0;
    top: 0;
    cursor: pointer;
  }
</style>
