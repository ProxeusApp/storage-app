<template>
  <div class="file-upload-container">
    <b-modal :id="'fileUploadModal' + _uid"
             modal-class="modal-file-upload"
             ref="fileUploadModal"
             :title="newFile ? newFile.name : $t('filebrowser.fileupload.upload_file', 'Upload File')"
             :lazy="true"
             :visible="modal"
             size="lg"
             :ok-title="$t('filebrowser.fileupload.upload', 'Upload')"
             :ok-disabled="formInvalid"
             :cancel-title="$t('generic.button.cancel', 'Cancel')"
             :busy="uploading || costEstimationInProgress"
             :no-close-on-backdrop="true"
             @show="onShow"
             @ok="upload"
             @hidden="close">

      <div v-if="uploading" class="row spinner-wrapper">
        <spinner background="transparent"></spinner>
      </div>
      <div v-else class="row">
        <div class="col-md-4 pr-0">
          <div class="thumbnail-container">
            <div class="upload-thumbnail image-holder">
              <img :src="thumbnailFile" width="100%" v-if="thumbnailFile">
              <file-icon v-if="thumbnailFile === undefined && newFile" :file="newFile" fontSize="5rem"
                         style="width: 100%; height: auto; cursor:inherit;"></file-icon>
              <div class="upload-thumbnail--actions">
                <!--<thumbnail-upload-button v-on:thumbnailChanged="thumbnailChanged"></thumbnail-upload-button>-->
                <thumbnail-upload-button :costEstimationInProgress="costEstimationInProgress" v-on:thumbnailChanged="thumbnailChanged"></thumbnail-upload-button>
              </div>
            </div>
          </div>
        </div>

        <div class="col-md-8">
          <div class="form-group">
            <label for="durationInput">{{ $t('filebrowser.fileupload.duration', 'Storage duration') }}</label>
            <div class="input-group">
              <input type="number"
                     id="durationInput"
                     v-model="durationHandler"
                     class="form-control w-75"
                     aria-label="Text input with dropdown button">
              <select class="custom-select w-25" id="inputGroupSelect01" v-model="durationType">
                <option
                  value="days">{{ $t('filebrowser.fileupload.duration.days', {days: durationHandler}, durationHandler) }}
                </option>
                <option
                  value="weeks">{{ $t('filebrowser.fileupload.duration.weeks', {weeks: durationHandler}, durationHandler) }}
                </option>
                <option
                  value="months">{{ $t('filebrowser.fileupload.duration.months', {months: durationHandler}, durationHandler) }}
                </option>
                <option
                  value="years">{{ $t('filebrowser.fileupload.duration.years', {years: durationHandler}, durationHandler) }}
                </option>
              </select>
            </div>
            <small v-if="isDurationValid" class="text-muted">{{ $t('filebrowser.fileupload.help_text_duration', 'Your file will be deleted after this period') }}</small>
            <small v-else-if="!isDurationValid && durationInDays < 1" class="text-danger">{{ $t('filebrowser.fileupload.error_text_duration_zero', 'The duration must be at least 1 day.') }}</small>
            <small v-else class="text-danger">{{ $t('filebrowser.fileupload.error_text_duration', 'You can store your file up to {maxDuration} days', { maxDuration: maxDuration}) }}</small>
          </div>

          <div class="form-group">
            <label
              for="selectStorageProvider">{{ $t('filebrowser.fileupload.select_storage_provider.label', 'Select storage provider') }}
            </label>
            <multiselect v-model="storageProviderHandler"
                         :options="storageProviders"
                         :multiple="false"
                         track-by="address"
                         label="name"
                         id="selectStorageProvider"
                         :searchable="false"
                         :closeOnSelect="true"
                         :show-labels="false"
                         :placeholder="$t('filebrowser.fileupload.select_storage_provider.placeholder', 'Select storage provider')">
              <template slot="singleLabel" slot-scope="props">
                <div class="d-flex flex-row align-items-center">
                  <div class="d-flex flex-column">
                    <div>{{ props.option.name }}</div>
                  </div>
                  <div class="ml-auto xes-price">{{ props.option.priceTotal | weiToXes }} XES</div>
                </div>
              </template>
              <template slot="option" slot-scope="props">
                <div class="d-flex flex-row align-items-center">
                  <div class="d-flex flex-column">
                    <div>{{ props.option.name }}</div>
                  </div>
                  <div class="ml-auto xes-price">{{ props.option.priceTotal | weiToXes }} XES</div>
                </div>
              </template>
            </multiselect>
            <small class="text-muted">{{ $t('filebrowser.fileupload.help_text_sp', 'Help Text') }}</small>
          </div>
          <div class="form-group">
            <label>{{ $t( 'filebrowser.fileupload.signatures_required', 'Signature(s) required?') }}
            </label>
            <div class="custom-control custom-radio mb-1">
              <input class="custom-control-input" type="radio" name="signersRequiredRadios" id="signersNotRequired"
                     v-bind:value="false" v-model="signersRequired">
              <label class="custom-control-label pl-2" for="signersNotRequired">{{ $t('general.no', 'No') }}</label>
            </div>
            <div class="custom-control custom-radio mb-1">
              <input class="custom-control-input" type="radio" name="signersRequiredRadios" id="signersRequired"
                     v-bind:value="true" v-model="signersRequired">
              <label class="custom-control-label pl-2" for="signersRequired">{{ $t('general.yes', 'Yes') }}</label>
            </div>
          </div>
          <div v-if="signersRequired" class="form-group">
            <label>{{ $t('filebrowser.fileupload.signes_known','Are the signers known?') }}</label>
            <div class="custom-control custom-radio mb-1">
              <input class="custom-control-input" type="radio" name="signersKnownRadios" id="signersNotKnown"
                     v-bind:value="false" v-model="signersDefined">
              <label class="custom-control-label pl-2" for="signersNotKnown">{{ $t('general.no', 'No') }}</label>
            </div>
            <div class="custom-control custom-radio mb-1">
              <input class="custom-control-input" type="radio" name="signersKnownRadios" id="signersKnown"
                     v-bind:value="true" v-model="signersDefined">
              <label class="custom-control-label pl-2" for="signersKnown">{{ $t('general.yes', 'Yes') }}</label>
            </div>
          </div>
          <div v-if="signersRequired && signersDefined" class="form-group">
            <label>{{ $t('filebrowser.fileupload.select_signers.label', 'Select signers') }}</label>
            <multiselect v-model="signers"
                         :options="addressesWithPGPKey"
                         :multiple="true"
                         track-by="address"
                         label="name"
                         :taggable="true"
                         @tag="tag"
                         :hide-selected="true"
                         :closeOnSelect="true"
                         :show-labels="false"
                         :placeholder="$t('filebrowser.fileupload.select_signers.placeholder', 'Select signers')"/>
            <small
              class="text-muted">{{ $t( 'filebrowser.fileupload.help_text', 'You can select only contacts with an attached PGP key.') }}
            </small>
          </div>
          <div v-if="signersRequired && !signersDefined" class="form-group">
            <label for="signersCount">{{ $t( 'filebrowser.fileupload.number_signatories', 'Number of Signatures') }} </label>
            <input class="form-control" v-model="undefinedSigners" id="signersCount" type="text">
            <small :class="{'text-muted' : !undefinedSignatoriesInvalid, 'text-danger' : undefinedSignatoriesInvalid}">
              {{ $t( 'filebrowser.fileupload.enter_number_signatories', 'Enter the number of signers required to sign the document.') }}
            </small>
          </div>
          <costs v-if="storageProviderHandler"
                 class="mt-3"
                 :storageProviders="[storageProviderHandler]"
                 :gasEstimate="gasEstimate"/>
        </div>
      </div>
    </b-modal>
  </div>
</template>

<script>
import Spinner from '@/components/Spinner'
import Multiselect from 'vue-multiselect'
import ThumbnailUploadButton from '@/components/ThumbnailUploadButton'
import BaseModal from '@/components/Modal/BaseModal'
import moment from 'moment'
import web3Utils from 'web3-utils'
import Costs from '@/components/Costs'

import filepdf from '../../assets/file-pdf.svg'
import filedoc from '../../assets/file-document.svg'
import fileimg from '../../assets/file-image.svg'
import filevid from '../../assets/file-video.svg'
import filewrd from '../../assets/file-word.svg'
import filechr from '../../assets/file-chart.svg'
import fileg from '../../assets/file-generic.svg'

export default {
  name: 'file-upload',
  extends: BaseModal,
  props: ['modal', 'newFile'],
  components: {
    Multiselect,
    Spinner,
    ThumbnailUploadButton,
    Costs
  },
  data () {
    return {
      newFileExpiry: undefined,
      newFileThumbnail: undefined,
      newFileThumbnailUrl: undefined,
      uploading: false,
      fileHash: '',
      signersRequired: false,
      signersDefined: false,
      signers: [],
      undefinedSigners: '1',
      storageProvider: undefined,
      storageProviderQuotes: [],
      durationType: 'days',
      duration: 1,
      uploadTriggered: false,
      gasEstimate: undefined,
      maxUndefinedSignatories: 50,
      costEstimationTimeout: null,
      costEstimationInProgress: false
    }
  },
  watch: {
    async newFileThumbnail () {
      await this.updateCostEstimationSync(false, true) // call sync on thumbnail to make sure new file has been processed
      if (this.newFileThumbnail !== undefined) {
        let timestampSuffix = '?a=' + new Date().getTime() // url does not change so we suffix timestamp to trigger reload
        this.newFileThumbnailUrl = this.$store.getters.thumbnailSrc(this.fileHash) + timestampSuffix
      }
    },
    duration: function () {
      this.updateCostEstimation(false, true)
    },
    durationType: function () {
      this.updateCostEstimation(false, true)
    },
    signersRequired () {
      this.updateCostEstimation()
    },
    signersDefined () {
      this.updateCostEstimation()
    },
    signers () {
      this.updateCostEstimation()
    },
    undefinedSigners: function () {
      // in order to avoid quirky js errors undefinedSigners must be input.type = "text" and toString() must be called.
      this.undefinedSigners = this.undefinedSigners.toString().replace(/[^0-9]/g, '')
      if (this.undefinedSigners > 0) {
        this.updateCostEstimation(true)
      }
    },
    storageProviderHandler (newSPHandler, oldSPHandler) {
      if (this.storageProviderHandler && this.gasEstimate !== undefined && newSPHandler.address !== oldSPHandler.address) {
        this.updateCostEstimation(true)
      }
    }
  },
  filters: {
    weiToXes: function (val) {
      if (val === undefined) {
        return 0
      }

      let xes = web3Utils.fromWei(val.toString())
      return parseFloat(xes).toFixed(4)
    }
  },
  computed: {
    durationInDays () {
      switch (this.durationType) {
        case 'days':
        default:
          return this.duration
        case 'weeks':
          return moment.duration(this.duration, 'weeks').asDays()
        case 'months':
          return moment.duration(this.duration, 'months').asDays()
        case 'years':
          return moment.duration(this.duration, 'years').asDays()
      }
    },
    durationHandler: {
      get () {
        return this.duration
      },
      set (value) {
        this.duration = parseInt(value)
      }
    },
    insufficientEtherBalanceText () {
      return this.$t('filebrowser.fileupload.insufficientEtherBalance',
        'Your current balance of {ethBalance} ETH is insufficient for this transaction.  Please top up your Proxeus Wallet. Get more <a href="http://faucet.ropsten.be:3001/">Ether tokens</a>',
        { ethBalance: this.ethBalance })
    },
    undefinedSignatoriesInvalid () {
      return this.signersRequired && !this.signersDefined &&
        (this.undefinedSigners > this.maxUndefinedSignatories || this.undefinedSigners === '' || this.undefinedSigners === '0') // '0' as string
    },
    formInvalid () {
      return this.storageProviderHandler === undefined ||
        (this.signersRequired && this.signersDefined && this.signers.length === 0) ||
        this.undefinedSignatoriesInvalid || !this.isDurationValid
    },
    uploadModalLoading: {
      get () {
        return this.$store.state.file.uploadModalLoading
      },
      set (loading) {
        this.$store.commit('SET_UPLOAD_MODAL_LOADING', loading)
      }
    },
    maxDuration () {
      return this.storageProviderHandler !== undefined ? this.storageProviderHandler.maxStorageDays : 0
    },
    isDurationValid () {
      return this.durationInDays > 0 && this.storageProviderHandler !== undefined && this.maxDuration >= this.durationInDays
    },
    maxFileSizeMB () {
      // only works in a single-spp environment, when having multiple spps this needs to be refactored.
      return this.storageProviderHandler !== undefined && this.storageProviderHandler.maxFileSizeByte !== undefined
        ? this.storageProviderHandler.maxFileSizeByte / 1000000 : 0
    },
    allowance () {
      return this.$store.state.wallet.allowance
    },
    xesBalance () {
      return this.$store.state.wallet.balance
    },
    ethBalance () {
      return this.$store.state.wallet.ethBalance
    },
    addressesWithPGPKey () {
      return this.$store.getters.addressesWithPGPKey
    },
    storageProviders () {
      return this.storageProviderQuotes.map((sp) => {
        return Object.assign(sp, this.$store.getters.activeStorageProviders.find(asp => {
          return asp.address === sp.provider.address
        }))
      })
    },
    storageProviderHandler: {
      get () {
        return this.storageProvider || this.storageProviders[0]
      },
      set (storageProvider) {
        this.storageProvider = storageProvider
      }
    },
    thumbnailFile () {
      if (this.newFile === undefined) {
        return ''
      }
      if (this.newFileThumbnailUrl === undefined) {
        const fileExtension = this.newFile.name.split('.').pop().toLowerCase()
        switch (fileExtension) {
          case 'pdf':
            return filepdf
          case 'rtf':
          case 'odt':
            return filedoc
          case 'doc':
          case 'docx':
            return filewrd
          case 'jpg':
          case 'jpeg':
          case 'png':
          case 'bmp':
          case 'gif':
          case 'tiff':
            return fileimg
          case 'avi':
          case 'mpg':
          case 'mov':
          case 'wmv':
          case 'mp4':
            return filevid
          case 'xls':
            return filechr
          default:
            return fileg
        }
      }
      return this.newFileThumbnailUrl
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
    }
  },
  methods: {
    async onShow (event) {
      if (this.gasEstimate === undefined) {
        this.uploadModalLoading = true
        event.preventDefault()
        await this.getQuotes()
        const response = await this.estimateGas()
        this.uploadModalLoading = false
        if (response.status === false) {
          switch (response.msg) {
            case 'gas required exceeds allowance or always failing transaction':
              const xesTransactionCost = this.$options.filters.weiToXes(this.storageProviderHandler.priceTotal)
              if (xesTransactionCost > parseFloat(this.xesBalance)) {
                this.$emit('insufficientXes', xesTransactionCost)
              } else if (xesTransactionCost > parseFloat(this.allowance)) {
                this.$emit('insufficientXesAllowance', xesTransactionCost)
              } else {
                this.$showNotification('general.notification.error.title', 'filebrowser.notify.fileRegistered', 'error')
              }
              break
            case 'file exceeds provider file size limit':
              this.showFileSizeNotification()
              break
            default:
              console.log(response.msg)
              this.$showNotification('filebrowser.notify.fileErrorGeneric.title', 'filebrowser.notify.fileErrorGeneric', 'error', {}, 5000)
          }
          this.close()
        } else {
          this.$refs.fileUploadModal.show()
        }
      }
    },
    async getQuotes () {
      if (this.modal === false) {
        return
      }
      const formData = this.prepareUpload()
      const response = await this.$store.dispatch('QUOTE', { formData })

      if (response && response.data && response.data.providers !== undefined && response.data.providers.length) {
        this.storageProviderQuotes = response.data.providers
        this.fileHash = response.data.fileHash
      }
    },
    async estimateGas () {
      if (this.modal === false) {
        return
      }
      const formData = this.prepareUpload()
      const response = await this.$store.dispatch('UPLOAD_ESTIMATE_GAS', { formData })

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }

      return response
    },
    updateCostEstimation (skipQuotes, skipGasEstimation) {
      if (this.costEstimationTimeout) {
        window.clearTimeout(this.costEstimationTimeout)
      }
      this.costEstimationInProgress = true
      this.costEstimationTimeout = window.setTimeout(async () => {
        if (!skipQuotes) {
          await this.getQuotes()
        }

        if (!skipGasEstimation && this.gasEstimate !== undefined) {
          await this.estimateGas()
        }
        this.costEstimationInProgress = false
      }, 800)
    },
    async updateCostEstimationSync (skipQuotes, skipGasEstimation) {
      this.costEstimationInProgress = true
      if (!skipQuotes) {
        await this.getQuotes()
      }

      if (!skipGasEstimation && this.gasEstimate !== undefined) {
        await this.estimateGas()
      }
      this.costEstimationInProgress = false
    },
    tag (searchQuery, id) {
      this.signers.push({ name: searchQuery, address: searchQuery })
    },
    prepareUpload () {
      let formData = new FormData()
      formData.append('file', this.newFile)

      if (this.newFileThumbnail) {
        formData.append('thumbnail', this.newFileThumbnail)
      }

      if (this.storageProviderHandler) {
        formData.append('providerAddress', this.storageProviderHandler.address)
      }

      if (this.signersRequired === true && this.signersDefined === true) {
        formData.append('definedSigners', JSON.stringify(this.signers))
      } else if (this.signersRequired === true && this.signersDefined === false) {
        formData.append('undefinedSigners', JSON.stringify(parseInt(this.undefinedSigners)))
      }

      if (this.durationInDays) {
        formData.append('duration', this.durationInDays)
      }

      return formData
    },
    async upload (evt) {
      evt.preventDefault()
      this.uploadTriggered = true
      this.uploading = true
      if (!this.storageProviderHandler) {
        this.$showNotification('filebrowser.notify.error', 'filebrowser.notify.storageProviderNotFound', 'error')
        this.uploading = false
        return
      }

      let formData = this.prepareUpload()

      formData.append('xesAmount', this.storageProviderHandler.priceTotal)

      const xesTransactionCost = this.$options.filters.weiToXes(this.storageProviderHandler.priceTotal)
      const response = await this.$store.dispatch('UPLOAD', formData)

      if (response.status === false && response.msg) {
        switch (response.msg) {
          case 'file already exists':
            this.$showNotification('general.notification.error.title', 'filebrowser.notify.fileRegistered', 'error')
            this.close()
            return
          case 'gas required exceeds allowance or always failing transaction':
            if (xesTransactionCost > parseFloat(this.allowance)) {
              this.$emit('insufficientXesAllowance', xesTransactionCost)
            } else {
              this.$showNotification('general.notification.error.title', 'filebrowser.notify.fileRegistered', 'error')
            }
            this.close()
            return
          case 'insufficient funds for gas * price + value':
            this.insufficientGasEstimationModal = true
            this.close()
            return
          case 'PGP public key missing':
            this.$showNotification('general.notification.error.title', 'filebrowser.notify.noPgpAttached', 'error')
            this.close()
            return
          case 'file exceeds file size limit':
            this.$showNotification('general.notification.error.title', 'filebrowser.notify.fileExceedsFileSizeLimit',
              'error')
            this.close()
            return
          default:
            this.$showNotification('general.notification.error.title', 'fileJS.upload.couldNotUpload', 'error')
            this.close()
            return
        }
      }
      this.$emit('uploaded')
      this.close()
    },
    thumbnailChanged (file) {
      this.newFileThumbnail = file
    },
    close () {
      // omit REMOVE_FILE_DISK_KEEP_META if upload was triggered, else cleanup
      // we keep meta because it may be an already uploaded file and don't want to remove thumb
      if (this.uploadTriggered !== true && this.fileHash) {
        this.$store.dispatch('REMOVE_FILE_DISK_KEEP_META', this.fileHash)
      }
      this.$emit('modalClosed', this.fileHash)
      this.reset()
    },
    reset () {
      this.newFileThumbnail = undefined
      this.newFileThumbnailUrl = undefined
      this.signers = []
      this.undefinedSigners = 1
      this.signersRequired = false
      this.signersDefined = false
      this.uploading = false
      this.durationType = 'days'
      this.duration = 1
      this.uploadTriggered = false
      this.fileHash = ''
      this.gasEstimate = undefined
    },
    showFileSizeNotification () {
      if (this.maxFileSizeMB) {
        this.$notify({
          title: this.$t('general.notification.error.title', 'Error'),
          text: this.$t('filebrowser.notify.fileExceedsFileSizeLimit', 'The upload is limited to a file size of {maxFileSize} MB', { maxFileSize: this.maxFileSizeMB }),
          type: 'error',
          duration: null
        })
      } else {
        this.$notify({
          title: this.$t('general.notification.error.title', 'Error'),
          text: this.$t('filebrowser.notify.fileExceedsFileSizeLimitGeneric', 'The File exceeds the filesize limit. Please select another file.'),
          type: 'error',
          duration: null
        })
      }
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  @media (min-width: 576px) {
    /deep/ .modal-dialog {
      max-width: 700px;
    }
  }

  .spinner-wrapper {
    min-height: 200px;
  }

  .receipt {
    border-bottom-right-radius: $border-radius;
    border-bottom-left-radius: $border-radius;
    border-top: 1px solid $gray-300;
  }

  .upload-placeholder {
    background: $light;
    height: 190px;
    width: 190px;
  }

  .thumbnail-container {
    background: $gray-300;
    padding: 5px;
    width: 200px;
  }

  .upload-thumbnail {
    width: 190px;
    height: 190px;
    overflow: hidden;
    min-width: 190px;
    cursor: pointer;
    position: relative;

    img {
      position: absolute;
      left: 50%;
      top: 50%;
      height: 100%;
      width: auto;
      -webkit-transform: translate(-50%, -50%);
      -ms-transform: translate(-50%, -50%);
      transform: translate(-50%, -50%);
    }

    .upload-thumbnail--actions {
      position: absolute;
      display: block;
      bottom: 5px;
      right: 5px;
      width: 100%;
      height: 100%;

      .btn {
        color: white;
        text-shadow: 1px 1px 10px rgba(0, 0, 0, 0.6);
      }
    }
  }
</style>
