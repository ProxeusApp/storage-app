<template>
  <div class="wallet-account">
    <!--<ul class="nav nav-tabs">-->
    <!--<li class="nav-item">-->
    <!--<a class="nav-link active" data-toggle="tab" href="#main" role="tab">Account</a>-->
    <!--</li>-->
    <!--<li class="nav-item">-->
    <!--<a class="nav-link" data-toggle="tab" href="#transactions" role="tab">Transactions</a>-->
    <!--</li>-->
    <!--<li class="nav-item">-->
    <!--<a class="nav-link" data-toggle="tab" href="#settings" role="tab">Settings</a>-->
    <!--</li>-->
    <!--</ul>-->
    <div class="tab-content" id="myTabContent" v-show="pane === 'overview'">
      <div class="tab-pane fade show active text-center px-3 pt-3 pb-2" id="main" role="tabpanel" aria-labelledby="home-tab">
        <h2 class="text-center wallet-headline text-truncate">{{ currentAccountName }}</h2>
        <div class="account-hash">
          <div class="form-group text-muted">
            <small>
              <strong class="hash-title">{{ $t('account.address', 'Address:') }}</strong>
              <a :href="`${etherscanUrl}/address/${address}`"
                 class="hash ml-1"
                 target="_blank"
                 onclick="window.openInBrowser(event, this);"
                 v-tooltip="$t('wallet.deposit.checkWallet', 'View wallet on etherscan')">
                {{ address }} <i class="mdi mdi-launch"></i>
              </a>
            </small>
          </div>
          <div class="accounting">
            <div class="accounting-balance bg-light py-3">
              <h5>{{ $t('account.ballance', 'Wallet balance') }}</h5>
              <div class="balance d-flex flex-column justify-content-around bg-light">
                <div>
                  <span class="xes xes-nr mr-1">{{ balance }}</span> <span class="xes mr-1">XES</span>
                  <br>
                  <span class="text-muted"><small>{{$t('account.total.xesBallance', 'Your XES token balance')}}</small></span>
                </div>
                <div class="mt-2">
                  <span class="eth eth-nr mr-1">{{ ethBalance }}</span> <span class="eth mr-1">ETH</span>
                  <br>
                  <span class="text-muted"><small>{{ $t('account.total.ethBallance', 'Your Ether balance') }}</small></span>
                </div>
                <div class="mt-2">
                  <button @click="setActivePane('deposit')"
                          class="btn btn-primary">{{ $t('account.wallet.deposit', 'Deposit') }}
                  </button>
                  <button @click="setActivePane('send')"
                          class="btn btn-primary ml-2">{{ $t('account.wallet.send', 'Send') }}
                  </button>
                </div>
              </div>
            </div>
            <div class="accounting-allowance bg-light py-3">
              <h5>{{ $t('account.wallet.allowance', 'PSPP Allowance') }}</h5>
              <div class="allowance d-flex flex-column justify-content-end bg-light">
                <div>
                  <span class="xes xes-nr mr-1">{{ allowance }}</span> <span class="xes mr-1">XES</span>
                  <i class="mdi md-15 mdi-information-outline icon"
                     v-tooltip="$t('account.wallet.tokenUsage', `These tokens will be used to upload and share files.`)"></i>
                </div>
                <div class="mt-2">
                  <button class="btn btn-primary" @click="setActivePane('approve')">{{ $t('account.wallet.changeAllowance',
                    `Change allowance`) }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <deposit v-if="pane === 'deposit'"></deposit>
    <send v-if="pane === 'send'" @modalClosed="$emit('modalClosed')"></send>
    <div v-if="pane === 'approve'" class="pt-0 pb-2">
      <div class="border-bottom">
        <button class="btn btn-link py-2" @click="setActivePane('overview')">
          <span class="mdi mdi-chevron-left md-24 d-inline-block border-right pr-1"></span>
          <span class="d-inline-block pl-1">{{ $t('account.wallet.backAccountOverview', 'Back to account overview') }}</span>
        </button>
      </div>
      <div class="balance-approve text-center mt-3" v-show="!approving">
        <h1>{{ $t('account.wallet.setAllowance', 'Set Allowance') }}</h1>
        <p class="text-muted">{{ $t('account.wallet.approveXes', `Approve XES tokens for covering file upload/signing/sharing etc.`) }}</p>
        <form @submit.prevent="approve">
          <span class="xes xes-nr">{{ approvalValue }}</span> <span class="xes">/</span>
          <span class="xes xes-nr mr-1">{{ sliderMax }}</span> <span class="xes">XES</span>
          <div class="form-group">
            <range-slider
              class="slider w-50 mt-3"
              min="0"
              :max="sliderMax"
              step="1"
              v-model="approvalValue">
            </range-slider>
          </div>
          <costs v-if="gasEstimate !== undefined"
                 class="mx-auto mb-2 text-left"
                 :gasEstimate="gasEstimate"/>
          <button type="submit" class="btn btn-primary mt-1"
                  :disabled="approvalValue === allowance">
            {{ $t('account.wallet.approve', 'Approve') }} {{ approvalValue }} XES
          </button>
        </form>
      </div>
      <div class="balance-approve text-center mt-3" v-show="approving">
        <h2>{{ $t('account.wallet.approvingTransaction', 'Approving Transaction…') }}</h2>
        <spinner background="transparent" style="position: relative;"></spinner>
      </div>
    </div>
  </div>
</template>

<script>
import RangeSlider from 'vue-range-slider'
import Spinner from '../../components/Spinner'
import web3utils from 'web3-utils'
import Send from './Send'
import Deposit from './Deposit'
import Costs from '@/components/Costs'

export default {
  name: 'account',
  components: {
    Deposit,
    RangeSlider,
    Send,
    Spinner,
    Costs
  },
  data () {
    return {
      approvalValue: 0,
      gasEstimate: undefined,
      gasEstimateTimeout: null
    }
  },
  watch: {
    // set value to current allowance in watcher because allowance is pushed async from backend-api
    allowance: function (allowance) {
      this.approvalValue = allowance
    },
    pane () {
      if (this.pane === 'approve') {
        this.estimateGas()
      } else {
        this.gasEstimate = undefined
      }
    },
    approvalValue () {
      if (this.pane === 'approve') {
        clearTimeout(this.gasEstimateTimeout)
        this.gasEstimateTimeout = setTimeout(() => {
          this.estimateGas()
        }, 500)
      }
    }
  },
  computed: {
    allowance () {
      return this.$store.state.wallet.allowance
    },
    approving () {
      return this.$store.getters.isApproving
    },
    balance () {
      return this.$store.state.wallet.balance
    },
    ethBalance () {
      return this.$store.state.wallet.ethBalance
    },
    balanceToFloat () {
      return parseFloat(this.balance)
    },
    sliderMax () {
      return Math.floor(parseFloat(this.balance))
    },
    address () {
      return this.$store.state.wallet.currentAddress
    },
    currentAccountName () {
      return this.$store.getters.currentAccount !== null ? this.$store.getters.currentAccount.name : ''
    },
    insufficientGasEstimationModal: {
      get () {
        return this.$store.state.notification.insufficientGasEstimationModal
      },
      set (showModal) {
        this.$store.commit('SET_INSUFFICIENT_GAS_MODAL', showModal)
      }
    },
    pane () {
      return this.$store.state.wallet.activePane
    },
    etherscanUrl () {
      return this.$store.getters.etherscanUrl
    }
  },
  methods: {
    setActivePane (pane) {
      this.$store.commit('UPDATE_ACTIVE_WALLET_PANE', pane)
    },
    async estimateGas () {
      const response = await this.$store.dispatch('APPROVE_ESTIMATE_GAS', { xesValue: web3utils.toWei(this.approvalValue + '') })

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    },
    async approve () {
      const response = await this.$store.dispatch('APPROVE', { xesValue: web3utils.toWei(this.approvalValue + '') })
      if (response.status === false && response.msg) {
        switch (response.msg) {
          case 'insufficient funds for gas * price + value':
            this.insufficientGasEstimationModal = true
            break
          default:
            this.$notify({
              title: this.$t('fileJS.transaction_queue.approve_xes.error', 'Could not change allowance.'),
              type: 'warn'
            })
            break
        }
      }
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";
  @import '~vue-range-slider/dist/vue-range-slider.css';

  /deep/ .range-slider-fill {
    background-color: $info;
    height: 10px;
    border-radius: $border-radius;
  }

  /deep/ .range-slider-rail {
    height: 10px;
    border-radius: $border-radius;
  }

  .account-hash {
    word-wrap: break-word;
  }

  .nav-tabs {
    background: $info;

    .nav-link {
      color: white;

      &.active {
        color: $primary;
      }
    }
  }

  .xes {
    font-size: 1.5rem;
    font-weight: 100;

    &.xes-nr {
      font-weight: 400;
    }
  }

  .eth {
    font-size: 1.5rem;
    font-weight: 100;

    &.eth-nr {
      font-weight: 400;
    }
  }

  .hash {
    max-width: 100%;
    display: inline-block;
    text-overflow: ellipsis;
    overflow: hidden;
    white-space: nowrap;
    color: inherit;
  }

  .hash-title {
    overflow: hidden;
    display: inline-block;
  }

  .accounting {
    display: grid;
    grid-auto-columns: 300px;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    grid-gap: 1rem;
  }

  .accounting-balance,
  .accounting-allowance {
    border-radius: $border-radius;
  }

  .btn-file-input {
    position: absolute;
    font-size: 50px;
    opacity: 0;
    right: 0;
    top: 0;
  }

  .icon {
    position: relative;
    bottom: 0.2rem;
  }

  .balance,
  .allowance,
  .balance-approve {
    color: $primary;
  }

  .wallet-headline {
    font-size: 1.6rem;
  }

  .costs {
    width: 60%;
  }
</style>
