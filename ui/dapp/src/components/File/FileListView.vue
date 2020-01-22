<template>
  <div class="container-list mb-2" :class="{syncing: isFileSyncing, removing: isFileRemoving}" v-on:click="showPreview">
    <div class="file w-100">
      <div class="file-container p-1 d-flex flex-row align-items-center"
           :class="{'bg-alert-warning': file.aboutToExpire && !isFileSyncing, 'bg-alert-danger': file.inGracePeriod && !isFileSyncing}">
        <file-icon :file="file" fontSize="3rem" :loading="isFileSyncing" @showPreview="showPreview"></file-icon>
        <template v-if="isFileSyncing">
          <div class="file--meta flex-fill w-50 pl-3">
            <span class="file-hash flex-col-truncate file-meta-row" v-on:click.stop>
              <span>{{ $t('fileinfo.syncing', 'Syncing…') }}</span>
            </span>
          </div>
        </template>
        <template v-else>
          <div class="file--meta flex-fill w-50 pl-3">
            <span class="file-hash flex-col-truncate file-meta-row" v-on:click.stop>
              <span>{{ file.filename }}</span>
              <i v-if="file.aboutToExpire"
                 class="mdi md-18 mdi-alert-circle d-inline-block align-bottom ml-1"
                 v-tooltip.right="$t('fileinfo.alert.aboutToExpire', 'This file is going to expire {expiryFromNow}.', { expiryFromNow: this.fileExpiryFromNow })"></i>
              <i v-if="file.inGracePeriod"
                 class="mdi md-18 mdi-alert-circle d-inline-block align-bottom ml-1"
                 v-tooltip.right="$t('fileinfo.alert.inGracePeriod', 'This file has expired and is going to be deleted from the Storage Provider {gracePeriodEndFromNow}.', { gracePeriodEndFromNow: this.fileGracePeriodEndFromNow })"></i>
            </span>
          </div>
          <div v-if="isFileRemoving === false" class="file--status text-muted w-25 flex-fill d-flex flex-row px-3">
            <div class="signing-status" v-if="getFileSignStatus === FILE_CONSTANTS.NO_SIGNERS_REQUIRED">
              <i class="icon mdi md-18 mdi-verified d-inline-block ml-auto"></i>
              <small>{{ $t('fileinfo.no_signatures.required', 'No signatures required') }}</small>
            </div>
            <div class="signing-status" v-if="getFileSignStatus === FILE_CONSTANTS.SIGNED">
              <i class="icon mdi md-18 mdi-verified d-inline-block ml-auto"></i>
              <small>{{ $t('filegridview.dropdown.signed', 'Signed') }}</small>
            </div>
            <div class="signing-status" v-if="getFileSignStatus === FILE_CONSTANTS.UNSIGNED">
              <i class="icon mdi md-18 mdi-alert-circle-outline d-inline-block ml-auto"></i>
              <small v-tooltip="getMissingSignersInfo">{{ $t('filegridview.dropdown.unsigned', 'Unsigned') }}</small>
            </div>
            <div class="signing-status">
              <!--<i class="icon mdi md-18 mdi-alert-circle-outline d-inline-block ml-auto"></i>-->
              <!--<small v-tooltip="getMissingSignersInfo">{{ $t('filegridview.dropdown.unsigned', 'Unsigned') }}</small>-->
            </div>
          </div>
          <div v-else class="status-removing mr-3 d-flex flex-row align-items-center">
            <tiny-spinner></tiny-spinner>
            <div class="text-muted">
              <small>{{ $t('fileJS.transaction_queue.removing_file', 'Removing file')}}</small>
            </div>
          </div>
          <file-actions class="mx-3"
                        v-if="isFileRemoving === false && isFileExpired === false"
                        @download="download"
                        @showInfo="showPreview"
                        @sharePrompt="sharePrompt"
                        @signRequestPrompt="signRequestPrompt"
                        @unsharePrompt="unsharePrompt"
                        @removeFile="removeFile"
                        @removeFilePrompt="removeFilePrompt"
                        @removeFileLocalPrompt="removeFileLocalPrompt"
                        :file="file"></file-actions>
        </template>
      </div>
    </div>
    <file-preview :file="file"
                  :modal="filePreviewModal"
                  @download="download"
                  @sharePrompt="sharePrompt"
                  @signRequestPrompt="signRequestPrompt"
                  @unsharePrompt="unsharePrompt"
                  @removeFile="removeFile"
                  @removeFilePrompt="removeFilePrompt"
                  @modalClosed="filePreviewModal = false"></file-preview>
    <remove-file-warning :file="file"
                         :modal="removeFileWarningModal"
                         @removeFile="removeFile"
                         @modalClosed="removeFileWarningModal = false"></remove-file-warning>
    <remove-file-local-warning :file="file"
                               :modal="removeFileLocalWarningModal"
                               @removeFileLocal="removeFileLocal"
                               @modalClosed="removeFileLocalWarningModal = false"></remove-file-local-warning>
    <share-file :file="file"
                :modal="shareFileModal"
                @modalClosed="shareFileModal = false"></share-file>
    <sign-request-file :file="file"
                       :modal="signRequestModal"
                       @modalClosed="signRequestModal = false"></sign-request-file>
    <unshare-file :file="file"
                  :modal="unshareFileModal"
                  @modalClosed="unshareFileModal = false"></unshare-file>
  </div>
</template>

<script>
import FileBaseView from './FileBaseView'

export default {
  name: 'file-list-view',
  extends: FileBaseView
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .container-list {
    position: relative;

    &:hover {
      .file--actions {
        display: block;
      }
    }
  }

  .signing-status {
    vertical-align: middle;
    display: flex;
  }

  .file {
    transition: all 300ms ease;
    position: relative;

    .dropdown-menu {
      z-index: 1110;
    }

    .file-container {
      cursor: pointer;
      border-radius: $border-radius;
      background: $light;

      .badge {
        margin-left: 0.5em;
      }

      &:hover {
        .file--actions {
          display: block;
        }
      }

      .syncing & {
        cursor: auto;
      }

      .removing & {
        background: darken($gray-100, 4%);
      }
    }

    &:hover {
      .file-container {
        background: darken($light, 2%);
      }
    }

    .file-hash {
      font-size: 0.8rem;
      color: $gray-700;
      display: block;
    }

    .flex-col-truncate {
      overflow: hidden;
      min-width: 0;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .icon {
      display: inline-block;
      width: 25px;
    }
  }
</style>
