<template>
  <div class="create">
    <spinner v-if="creatingAccount" background="transparent" style="position: relative;"></spinner>
    <div v-else>
      <h5 class="login-text-hint">{{ $t('loginfullscreen.createNewAccount', 'Create a new account') }}</h5>
      <form @submit.prevent="createNewWalletAccount">
        <div class="form-group w-100">
          <input type="text" class="form-control" v-model="newAccountName" ref="walletName" tabindex="1"
                 :placeholder="$t('loginfullscreen.wallet.accountName', 'Wallet name')"/>
          <password id="setPasswordInputCreateWallet"
                    :inputTabindex="2"
                    :placeholder="$t('loginfullscreen.placeholder.password', 'Password')" :secureLength="6"
                    :toggle="true" defaultClass="form-control mt-3" @score="updateScore" @warning="updateWarning"
                    v-model="newAccountPassword"/>
          <small v-if="showPasswordWarning" class="auth-error text-danger">{{ newAccountPasswordWarning }}</small>
          <small v-else-if="newAccountPasswordScore > 1"
                 class="text-success">{{ $t('password.strength.strongpassword', 'You are using a safe password') }}
          </small>
          <small v-else class="text-muted mb-3">{{ $t('loginfullscreen.useSafePassword', 'Use a safe password.') }}</small>
          <input type="password" class="form-control mt-3" tabindex="3" v-model="newAccountPasswordRetype"
                 id="newAccountPasswordRetype"
                 name="newAccountPasswordRetype"
                 :placeholder="$t('loginfullscreen.placeholder.reTypePassword', 'Retype password')">
        </div>
        <div class="auth-error" v-if="authError">
          <span class="text-danger">{{ authError }}</span>
        </div>
        <div class="mt-3">
          <button type="submit"
                  class="btn btn-primary"
                  tabindex="4"
                  :disabled="createAccountDisabled">{{ $t('loginfullscreen.button.createAccount', 'Create Account') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import Spinner from '@/components/Spinner'
import Password from '@/components/StrengthMeter.vue'

export default {
  name: 'create-wallet',
  components: {
    Spinner,
    Password
  },
  data () {
    return {
      newAccountName: '',
      newAccountPassword: '',
      newAccountPasswordRetype: '',
      newAccountPasswordScore: 0,
      newAccountPasswordWarning: ''
    }
  },
  watch: {
    loginTabIndex: function (index) {
      if (index === 1) {
        this.$refs.walletName.focus()
      }
    }
  },
  computed: {
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
    loginTabIndex () {
      return this.$store.state.wallet.loginTabIndex
    },
    creatingAccount () {
      return this.$store.state.wallet.creatingAccount
    },
    createAccountDisabled () {
      return this.newAccountPassword === '' || this.newAccountPasswordRetype !== this.newAccountPassword ||
        this.newAccountPasswordScore < 2
    },
    showPasswordWarning () {
      return this.newAccountPassword.length > 0 && this.newAccountPasswordWarning !== '' &&
        this.newAccountPasswordScore < 2
    }
  },
  methods: {
    async createNewWalletAccount () {
      try {
        await this.$store.dispatch('CREATE_ACCOUNT', {
          name: this.newAccountName,
          password: this.newAccountPassword
        })
      } catch (e) {
        console.log(e)
      }
    },
    updateScore (score) {
      this.newAccountPasswordScore = score
    },
    updateWarning (warning) {
      this.newAccountPasswordWarning = warning
    }
  }
}
</script>

<style scoped>
  .btn.btn-primary {
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
