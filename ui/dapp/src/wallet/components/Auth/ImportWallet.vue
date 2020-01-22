<template>
  <div class="import">
    <spinner v-if="authenticating" background="transparent" style="position: relative;"></spinner>
    <div v-else>
      <h5 class="login-text-hint">{{ $t('loginfullscreen.importAccount', 'Import an account') }}</h5>
      <!--PSS-147: Disable ETH Private-Key import and only allow keystore import-->
      <div v-if="importType === 'pk'">
        <div class="form-group">
          <input type="text" class="form-control" v-model="accountName" ref="walletName" tabindex="2"
                 :placeholder="$t('loginfullscreen.wallet.accountName', 'Wallet name')">
        </div>
        <div class="form-group">
          <label
            for="formGroupExampleInput">{{ $t('loginfullscreen.wallet.ethereumPrivateKey', 'Ethereum Account Private Key') }}
          </label>
          <input type="text" class="form-control" autocomplete="new-password" id="formGroupExampleInput" tabindex="3"
                 placeholder="478FA0086D67DCB3DGPW2E1238…"
                 v-model="privateKey">
        </div>
        <div class="form-group">
          <label for="setPasswordInputImportWallet">{{ $t('loginfullscreen.wallet.password', 'Set Password') }}</label>
          <password id="setPasswordInputImportWallet" :placeholder="$t('loginfullscreen.password', 'Password')"
                    :inputTabindex="4"
                    :secureLength="6"
                    :toggle="true" defaultClass="form-control" @score="updateScore" @warning="updateWarning"
                    v-model="password"/>
          <small v-if="showPasswordWarning" class="auth-error text-danger">{{ passwordWarning }}</small>
          <small v-else-if="passwordScore > 1"
                 class="text-success">{{ $t('password.strength.strongpassword', 'You are using a safe password') }}
          </small>
          <small v-else class="text-muted mb-3">{{ $t('loginfullscreen.useSafePassword', 'Use a safe password.') }}</small>
        </div>
        <div class="form-group">
          <label for="retypePasswordInput">{{ $t('loginfullscreen.placeholder.reTypePassword', 'Retype password') }} </label>
          <input type="password" id="retypePasswordInput" class="form-control" v-model="passwordRetype" name="reTypePw" tabindex="5"
                 :placeholder="$t('loginfullscreen.placeholder.reTypePassword', 'Retype password')">
        </div>
        <div class="auth-error mb-3" v-if="authError">
          <span class="text-danger">{{ authError }}</span>
        </div>
        <button class="btn btn-primary" tabindex="9" @click="importPrivateKey"
                :disabled="importPKDisabled">{{ $t('loginfullscreen.button.import', 'Import') }}
        </button>
      </div>
      <div v-if="importType === 'file'">
        <form enctype="multipart/form-data" novalidate @submit.prevent class="text-center">
          <div class="form-group">
            <div class="btn btn-primary text-center">
              <span>{{ $t('loginfullscreen.import.file.button.label', 'Upload Wallet File') }}</span>
              <input class="btn-file-input" type="file" name="file" tabindex="10"
                     @change="keystoreFileChanged($event.target.files)"/>
            </div>
          </div>
          <template v-if="keystoreFile">
            <div class="form-group">
              <label for="keystore-password" class="w-100 text-truncate">
                {{ $t('loginfullscreen.import.passwordFor', 'Password for keystore:') }} {{ keystoreFile.name }}
              </label>
              <input type="password" id="keystore-password" class="form-control" v-model="keystorePassword" tabindex="11"
                     name="keystore-password" ref="passwordInput" :placeholder="$t('loginfullscreen.password', 'Password')">
            </div>
            <div class="auth-error mb-3" v-if="authError">
              <span class="text-danger">{{ authError }}</span>
            </div>
            <button class="btn btn-primary" tabindex="12" @click="importKeystoreFromFile"
                    :disabled="keystorePassword === ''">{{ $t('loginfullscreen.button.import', 'Import') }}
            </button>
          </template>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import Spinner from '@/components/Spinner'
import Password from '@/components/StrengthMeter.vue'

export default {
  name: 'import-wallet',
  data () {
    return {
      importType: 'file',
      accountName: '',
      privateKey: '',
      password: '',
      passwordRetype: '',
      keystoreFile: undefined,
      keystorePassword: '',
      passwordWarning: '',
      passwordScore: 0
    }
  },
  components: {
    Spinner,
    Password
  },
  computed: {
    currentAddress: {
      get () {
        return this.state.wallet.currentAddress
      },
      set (addr) {
        this.$store.commit('SET_CURRENT_ADDRESS', addr)
      }
    },
    authenticating: {
      get () {
        return this.$store.state.wallet.authenticating
      },
      set (flag) {
        this.$store.commit('SET_AUTHENTICATING', flag)
      }
    },
    authError: {
      get () {
        return this.$store.state.wallet.authError
      },
      set (flag) {
        this.$store.commit('SET_AUTH_ERROR', flag)
      }
    },
    importPKDisabled () {
      return (this.privateKey.length === 0 || this.password.length < 6 || this.passwordScore < 2 ||
        this.password !== this.passwordRetype)
    },
    showPasswordWarning () {
      return this.password.length > 0 && this.passwordWarning !== '' && this.passwordScore < 2
    }
  },
  methods: {
    keystoreFileChanged (fileList) {
      if (fileList[0]) {
        this.keystoreFile = fileList[0]
        this.$nextTick(() => {
          if (typeof this.$refs.passwordInput !== 'undefined') {
            this.$refs.passwordInput.focus()
          }
        })
      }
    },
    async importKeystoreFromFile () {
      await this.$store.dispatch('IMPORT_KEYSTORE_FROM_FILE', { walletFile: this.keystoreFile, password: this.keystorePassword })
      if (this.authError === '') {
        this.keystoreFile = undefined
        this.keystorePassword = ''
      }
    },
    async importPrivateKey () {
      try {
        let res = await this.$store.dispatch('IMPORT_PRIVATE_KEY', {
          accountName: this.accountName,
          privateKey: this.privateKey,
          password: this.password
        })

        if (res === false) {
          this.importType = 'pk' // todo: fix resetting importType to pk
        }
      } catch (e) {
        this.importType = 'pk' // todo: fix resetting importType to pk
      }
    },
    updateScore (score) {
      this.passwordScore = score
    },
    updateWarning (warning) {
      this.passwordWarning = warning
    }
  }
}
</script>

<style lang="scss" scoped>
  .btn.btn-primary {
    overflow: hidden;
    position: relative;
  }

  .btn-file-input {
    position: absolute;
    font-size: 50px;
    opacity: 0;
    right: 0;
    top: 0;
    cursor: pointer;
  }
</style>
