<template>
  <div class="container-grid file-grid-view-container" :class="{syncing: isFileSyncing, removing: isFileRemoving}">
    <div class="file">
      <div class="file-container p-2"
           :class="{'bg-alert-warning': file.aboutToExpire && !isFileSyncing, 'bg-alert-danger': file.inGracePeriod && !isFileSyncing}">
        <file-icon :file="file" fontSize="5rem" :loading="isFileSyncing" @showPreview="showPreview" class="preview-grid"></file-icon>
        <template v-if="isFileSyncing">
          <div class="file--meta pt-1">
            <div class="file-hash flex-col-truncate file-meta-row">
              <span>{{ $t('fileinfo.syncing', 'Syncing…') }}</span>
            </div>
          </div>
        </template>
        <template v-else>
          <div class="file--meta pt-1">
            <div class="file-hash flex-col-truncate file-meta-row">
              <i v-if="file.aboutToExpire"
                 class="mdi md-18 mdi-alert-circle d-inline-block align-bottom mr-1"
                 v-tooltip="$t('fileinfo.alert.aboutToExpire', 'This file is going to expire {expiryFromNow}.', { expiryFromNow: this.fileExpiryFromNow })"></i>
              <i v-if="file.inGracePeriod"
                 class="mdi md-18 mdi-alert-circle d-inline-block align-bottom mr-1"
                 v-tooltip="$t('fileinfo.alert.inGracePeriod', 'This file has expired and is going to be deleted from the Storage Provider {gracePeriodEndFromNow}.', { gracePeriodEndFromNow: this.fileGracePeriodEndFromNow })"></i>
              <span>{{ file.filename }}</span>
            </div>
            <div v-if="isFileRemoving === false" class="file--status text-muted w-100 pt-1">
              <div class="signing-status" v-if="getFileSignStatus === FILE_CONSTANTS.NO_SIGNERS_REQUIRED">
                <small class="d-inline-block">{{ $t('fileinfo.no_signatures.required', 'No signatures required') }}</small>
                <i class="icon mdi md-18 mdi-verified d-inline-block ml-auto"></i>
              </div>
              <div class="signing-status" v-if="getFileSignStatus === FILE_CONSTANTS.SIGNED">
                <small class="d-inline-block">{{ $t('filegridview.dropdown.signed', 'Signed') }}</small>
                <i class="icon mdi md-18 mdi-verified d-inline-block ml-auto"></i>
              </div>
              <div class="signing-status" v-if="getFileSignStatus === FILE_CONSTANTS.UNSIGNED">
                <small class="d-inline-block"
                       v-tooltip="getMissingSignersInfo">{{ $t('filegridview.dropdown.unsigned', 'Unsigned') }}
                </small>
                <i class="icon mdi md-18 mdi-alert-circle-outline d-inline-block ml-auto"></i>
              </div>
            </div>
            <div v-else class="status-removing d-flex flex-row align-items-center">
              <tiny-spinner></tiny-spinner>
              <div class="text-muted">
                <small>{{ $t('fileJS.transaction_queue.removing_file', 'Removing file')}}</small>
              </div>
            </div>
          </div>
          <file-actions class="overview" v-if="isFileRemoving === false && isFileExpired === false"
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
  name: 'file-grid-view',
  extends: FileBaseView
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .file-grid-view-container {
    /deep/ .file--actions.overview {
      background: transparent;
      position: absolute;
      z-index: 1040;
      bottom: 59px;
      right: 10px;

      .btn-group > button {
        position: absolute;
        display: none;
        right: 7px;
        top: -40px;
      }

      .btn {
        background: transparent;
        border: none;
        box-shadow: none;

        &:focus,
        &:active {
          border: none;
          outline: none;
          box-shadow: none;
          background: transparent;
        }

        .icon {
          color: white;
          text-shadow: 1px 1px 5px rgba(0, 0, 0, 1);
        }

        &:hover .icon {
          color: #eeeeee;
          text-shadow: 1px 1px 10px rgba(0, 0, 0, 0.9);
        }
      }
    }

    .file {
      transition: all 300ms ease;
      position: relative;

      .signing-status {
        vertical-align: middle;
        display: flex;
        flex-direction: row;

        small {
          font-size: 75%;
        }
      }

      .dropdown-menu {
        z-index: 1110;
      }

      .file-container {
        cursor: pointer;
        border-radius: $border-radius;
        background: $light;

        &:hover {
          background: darken($light, 2%);

          /deep/ .file--actions .btn-group > button {
            display: block;
          }
        }
      }

      .file-hash {
        color: $gray-700;
        font-size: 0.8rem;
      }

      .file-meta-row {
        max-width: 200px;
      }

      .flex-col-truncate {
        overflow: hidden;
        min-width: 0;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    &.container-grid {
      position: relative;

      &:hover {
        .file--actions {
          display: block;
        }
      }
    }

    &.syncing {
      .file {
        .file-container {
          cursor: auto;
        }
      }
    }

    &.removing {
      .file {
        .file-container {
          background: darken($gray-100, 4%);
        }
      }
    }
  }
</style>
