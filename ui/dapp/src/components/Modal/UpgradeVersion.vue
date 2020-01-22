<template>
  <b-modal :id="'upgradeModal' + _uid"
           :title="$t('general.important.update', 'Important Update')"
           :lazy="true"
           :visible="modal"
           :no-close-on-backdrop="lockModal"
           :no-close-on-esc="lockModal"
           :hide-header-close="lockModal"
           @hidden="$emit('modalClosed')">
    <p>{{ $t('modal.upgradeversion.newversionavailable', 'A new version of the Proxeus DApp is available.') }}</p>
    <p
      v-if="isBlocking">{{ $t('modal.upgradeversion.upgradereqired', 'To continue using this DApp you have to upgrade to the newest version.') }}</p>
    <div v-if="showProgress">
      <h5>{{ $t('modal.upgradeversion.upgrading', 'Upgrading') }}</h5>
      <b-progress show-progress :value="counter" :max="max" class="mb-3"></b-progress>
    </div>
    <p v-if="result !== undefined && result === 'failed'" v-html="failedMsg"></p>
    <p
      v-if="result !== undefined && result === 'ready'">{{ $t('modal.upgradeversion.upgradeDownloaded', 'The upgrade is ready to be applied. Apply now.') }}</p>
    <p
      v-if="result !== undefined && result === 'success'">{{ $t('modal.upgradeversion.upgradeApplied', 'The upgrade was applied. Please restart to complete the process.') }}</p>
    <div slot="modal-footer" class="flex-wrap">
      <div class="row flex-fill">
        <div class="col-12 col-lg-auto ml-lg-auto text-right">
          <button v-if="isInfo"
                  type="button" class="btn btn-secondary"
                  @click="$emit('modalClosed')">{{ $t('close', 'Close') }}
          </button>
          <button v-if="showUpgradeBtn" type="button" class="btn btn-primary ml-1"
                  :disabled="upgrading"
                  @click="upgrade">{{ $t('modal.upgradeversion.upgrade', 'Upgrade') }}
          </button>
          <button v-if="showApplyBtn" type="button" class="btn btn-primary ml-1"
                  :disabled="applying"
                  @click="apply">{{ $t('modal.upgradeversion.apply', 'Apply') }}
          </button>
          <!--<a class="btn btn-primary" onclick="window.openInBrowser(event, this);" href="https://drive.google.com/drive/u/0/folders/1i9qWcx-QXedlrTMjUR2uvr2rk-h2G4kw?ths=true" target="_blank" role="button">{{ $t('filebrowser.buttons.upgrade', 'Upgrade') }}</a>-->
        </div>
      </div>
    </div>
  </b-modal>
</template>

<script>
import BaseModal from '@/components/Modal/BaseModal'
import axios from 'axios'

export default {
  name: 'upgrade-version',
  extends: BaseModal,
  props: ['modal', 'version'],
  data () {
    return {
      counter: undefined,
      max: 100,
      result: undefined,
      applying: false
    }
  },
  computed: {
    isBlocking () {
      return this.version === undefined ? false : this.version.update === 'block'
    },
    isInfo () {
      return this.version === undefined ? false : this.version.update === 'info'
    },
    upgrading () {
      return this.counter !== undefined && this.counter < 100
    },
    failedMsg () {
      return this.$t('modal.upgradeversion.upgradeFailed',
        'The upgrade failed. Visit {link} to download the latest version.',
        { link: '<a href="http://www.proxeus.com/" target="_blank">www.proxeus.com</a>' })
    },
    showUpgradeBtn () {
      return (this.counter === undefined || this.counter < this.max)
    },
    showApplyBtn () {
      return !this.showUpgradeBtn && !this.finished
    },
    showProgress () {
      return this.counter !== undefined || (this.result !== undefined && this.result === 'success')
    },
    lockModal () {
      return this.isBlocking || this.upgrading || (this.result === 'ready' && !this.finished)
    },
    finished () {
      return this.result === 'success' || this.result === 'failed'
    }
  },
  methods: {
    async upgrade () {
      axios.post('/api/update/download').then(res => {
        if (res.status === 200) {
          this.counter = 100
          this.result = 'ready'
        }
      }, () => {
        this.result = 'failed'
      })
      let sleep = function (ms) {
        return new Promise(resolve => setTimeout(resolve, ms))
      }
      for (let i = 0; i < 100; i++) {
        if (this.result === undefined &&
          (this.counter === undefined || this.counter < 100)) {
          this.counter = i
          await sleep(500)
        }
      }
    },
    apply () {
      axios.post('/api/update/apply').then(res => {
        this.applying = true
        if (res.status === 200) {
          this.applying = false
          this.result = 'success'
        }
      }, () => {
        this.result = 'failed'
      })
    }
  }
}
</script>
