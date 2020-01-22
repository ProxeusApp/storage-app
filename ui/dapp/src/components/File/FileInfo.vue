<template>
  <div ref="filePreviewModal">
    <file-info-entry :label="$t('fileinfo.filename', 'File Name')" class="mt-0">
      {{ file.filename }}
    </file-info-entry>

    <file-info-entry :label="$t('fileinfo.filehash', 'File Hash')"
                     :help="$t('fileinfo.filename.encrypted', 'The file hash of the encrypted file.')">
      <button :title="file.id"
              @click="doCopy(file.id, false)" type="button"
              class="btn btn-badge btn-badge-lg badge-proxeus">
        {{ file.id }}
      </button>
    </file-info-entry>

    <file-info-entry :label="$t('fileinfo.expiry', 'Expiration date')"
                     :help="fileExpiryHelpText"
                     :infoClass="{'bg-alert-warning': file.aboutToExpire, 'bg-alert-danger': file.inGracePeriod || file.expired}">
      {{ fileExpiryDate }}
    </file-info-entry>

    <file-info-entry :label="$t('fileinfo.ownerTitle', 'Owner')">
      <button :title="file.owner.address"
              @click="doCopy(file.owner.address, true)" type="button"
              class="btn btn-badge mr-1 badge-proxeus">
        {{ ownerName }}
      </button>
    </file-info-entry>

    <file-info-entry :label="$t('fileinfo.read_access', 'Read Access')"
                     v-if="file.readAccess && file.readAccess.length > 0">
      <button v-for="address in file.readAccess"
              :key="address.address"
              :title="address.address"
              @click="doCopy(address.address, true)" type="button"
              class="btn btn-badge mr-1 badge-proxeus">
        {{ nameByAddress(address.address) }} <span class="badge"></span> <span
        class="sr-only">{{ $t('fileinfo.read_access.signer', 'signer') }}</span>
      </button>
      <span v-if="!file.readAccess">{{ $t('fileinfo.nobody', 'Nobody except owner') }}</span>
    </file-info-entry>

    <div v-if="file.fileType === FILE_CONSTANTS.UNDEFINED_SIGNERS">
      <file-info-entry v-if="file.undefinedSigners > 0"
                       :label="$t('fileinfo.read_access.signers', 'Signatories')"
                       :help="$t('fileinfo.signers.required', 'The signers required to sign the document')">
        <span>{{ file.undefinedSigners - file.undefinedSignersLeft || 0 }} / {{ file.undefinedSigners }} {{ $t('fileinfo.total.signers')}}</span>
        <button v-for="entry in getSignersList"
                :key="entry.signer.address"
                :title="entry.signer.address"
                @click="doCopy(entry.signer.address, true)"
                type="button"
                class="btn btn-badge mr-1" :class="entry.signed ? 'badge-proxeus' : 'badge-danger'">
          {{ nameByAddress(entry.signer.address) }}
          <i v-if="entry.signed" class="icon mdi mdi-verified ml-auto"></i>
          <i v-else class="icon mdi mdi-alert-circle-outline ml-auto"></i>
        </button>
      </file-info-entry>

      <file-info-entry v-else
                       :label="$t('fileinfo.read_access.signers', 'Signatories')"
                       :help="$t('fileinfo.signers.required', 'The signers required to sign the document')">
        <span>{{ $t('fileinfo.no_signatures.required', 'No signatures required') }}</span>
      </file-info-entry>

      <file-info-entry v-if="file.signers">
        <button v-for="signer in file.signers"
                :key="signer.address"
                :title="signer.name"
                @click="doCopy(signer.address, true)"
                type="button"
                class="btn btn-badge mr-1 badge-proxeus">
          {{ signer.name }}
          <i class="icon mdi mdi-verified ml-auto"></i>
        </button>
      </file-info-entry>
    </div>

    <file-info-entry v-if="file.fileType === FILE_CONSTANTS.DEFINED_SIGNERS"
                     :label="$t('fileinfo.read_access.signers', 'Signatories')"
                     :help="$t('fileinfo.signers.required', 'The signers required to sign the document')">
      <div v-if="getFileSignStatus === FILE_CONSTANTS.SIGNED || getFileSignStatus === FILE_CONSTANTS.UNSIGNED"
           class="signatories">
        <button v-for="entry in getSignersList"
                :key="entry.signer.address"
                :title="entry.signer.address"
                @click="doCopy(entry.signer.address, true)"
                type="button"
                class="btn btn-badge mr-1" :class="entry.signed ? 'badge-proxeus' : 'badge-danger'">
          {{ nameByAddress(entry.signer.address) }}
          <i v-if="entry.signed" class="icon mdi mdi-verified ml-auto"></i>
          <i v-else class="icon mdi mdi-alert-circle-outline ml-auto"></i>
        </button>
      </div>
      <span v-else>{{ $t('fileinfo.no_signatures.required', 'No signatures required') }}</span>
    </file-info-entry>
  </div>
</template>

<script>
import moment from 'moment'
import { mapState } from 'vuex'
import FileInfoEntry from './FileInfoEntry'
import FILE_CONSTANTS from '../../lib/FileConstants'

export default {
  name: 'file-info',
  components: {
    FileInfoEntry
  },
  props: {
    file: {
      type: Object,
      required: true
    }
  },
  data () {
    return {
      FILE_CONSTANTS: FILE_CONSTANTS
    }
  },
  computed: {
    ...mapState({
      addresses: state => state.address.addresses
    }),
    ownerName () {
      return this.$store.getters.nameByAddress(this.file.owner.address)
    },
    fileExpiryDate () {
      return moment(this.file.expiry).format('YYYY-MM-DD')
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
    fileExpiryFromNow () {
      return moment(this.file.expiry).fromNow()
    },
    fileGracePeriodEndFromNow () {
      return moment(this.file.expiry).add(this.file.graceSeconds, 'seconds').fromNow()
    },
    fileGracePeriod () {
      return moment.duration(this.file.graceSeconds, 'seconds').humanize()
    },
    fileExpiryHelpText () {
      if (this.file.aboutToExpire) {
        return this.$t('fileinfo.expiry.helpAboutToExpire', 'This file is going to expire {expiryFromNow}. After the file has expired, it will remain on the storage provider for a grace period of {gracePeriod}.', { expiryFromNow: this.fileExpiryFromNow, gracePeriod: this.fileGracePeriod })
      } else if (this.file.inGracePeriod) {
        return this.$t('fileinfo.expiry.helpInGracePeriod', 'This file has expired and is going to be deleted from the Storage Provider {gracePeriodEndFromNow}.', { gracePeriodEndFromNow: this.fileGracePeriodEndFromNow })
      } else if (this.file.expired) {
        return this.$t('fileinfo.expiry.helpExpired', 'This file has expired and was deleted from the Storage Provider.')
      } else {
        return this.$t('fileinfo.expiry.help', 'After the file has expired, it will remain on the storage provider for a grace period of {gracePeriod}.', { gracePeriod: this.fileGracePeriod })
      }
    }
  },
  methods: {
    nameByAddress (address) {
      return this.$store.getters.nameByAddress(address)
    },
    doCopy (hash, isAddress) {
      let container = this.$refs.filePreviewModal
      this.$copyText(hash, container).then(function (e) {
      }, function (e) {
      })
      this.$notify({
        title: isAddress ? this.$t('fileinfo.notify.copyAddressClipboard', 'Copied address to Clipboard') : this.$t(
          'fileinfo.notify.copyFileHashClipboard', 'Copied File Hash to Clipboard'),
        text: '',
        type: 'success',
        duration: 2000
      })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .alert {
    margin-top: 0;
    box-shadow: none;
    text-align: left;
  }

  .badge-proxeus {
    background-color: darken($info, 5%);
    color: $white;
  }

  .badge-danger {
    background: darken($light, 40%);
  }

  .btn-badge {
    overflow: hidden;
    margin-top: 3px;
    margin-bottom: 3px;
    min-width: 0;
    text-overflow: ellipsis;
    white-space: nowrap;
    word-wrap: break-word;
    max-width: 250px;
    padding: 1px 12px;
    vertical-align: middle;
    display: inline-block;

    .icon {
      font-size: 16px;
      display: inline;
    }

    &.btn-badge-lg {
      max-width: 100%;
    }
  }
</style>
