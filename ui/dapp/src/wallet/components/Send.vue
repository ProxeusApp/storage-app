<template>
  <div>
    <div class="pt-0" v-if="step === 'form'">
      <div class="border-bottom">
        <button class="btn btn-link py-2" @click="backToOverview">
          <span class="mdi mdi-chevron-left md-24 d-inline-block border-right pr-1"></span> <span
          class="text-muted d-inline-block pl-1">{{ $t('account.wallet.backAccountOverview', 'Back to account overview') }}</span>
        </button>
      </div>
      <div class="deposit pt-3 px-3 text-center">
        <h2>{{ $t('wallet.send.formTitle', 'Send XES/ETH')}}</h2>
        <div class="text-muted">{{ $t('wallet.send.warning', 'Only send {currency} to an Ethereum account address', {currency}) }}</div>
        <div class="wallet-box">
          <form class="form text-left my-3 px-4 py-2">
            <div class="py-2 text-center">
              <div class="custom-control custom-radio custom-control-inline">
                <input type="radio"
                       id="customRadioInline1"
                       name="customRadioInline1"
                       value="XES"
                       v-model="currency"
                       class="custom-control-input">
                <label class="custom-control-label pl-2 font-weight-bold" for="customRadioInline1">XES</label>
              </div>
              <div class="custom-control custom-radio custom-control-inline">
                <input type="radio"
                       id="customRadioInline2"
                       name="customRadioInline1"
                       class="custom-control-input"
                       v-model="currency"
                       value="ETH">
                <label class="custom-control-label pl-2 ml-2 font-weight-bold" for="customRadioInline2">ETH</label>
              </div>
            </div>
            <div class="form-group row mb-1 pt-2">
              <div class="col-sm-2">
                <label for="balance">{{ $t('account.balance', 'Balance')}}</label>
              </div>
              <div class="col-sm-10"><span id="balance">{{ balance }} {{ currency }}</span></div>
            </div>
            <div class="form-group row">
              <label for="selectToAddres" class="col-sm-2 col-form-label">To</label>
              <div class="col-sm-10">
                <multiselect v-model="toAddress"
                             :options="contacts"
                             :multiple="false"
                             track-by="address"
                             label="name"
                             required
                             id="selectToAddres"
                             :hide-selected="false"
                             :closeOnSelect="true"
                             :placeholder="$t('filegridview.share.selectContacts', 'Select a contact or enter an account address')"
                             :taggable="true"
                             @tag="setCustomAddress"/>
                <div class="text-danger" v-if="!addressEmptyOrValid">
                  <small>{{ $t('wallet.send.addressInvalid', 'Enter a valid Ethereum Address')}}</small>
                </div>
              </div>
            </div>
            <div class="form-group mb-2 row">
              <label for="inputAmount" class="col-sm-2 col-form-label">{{ $t('wallet.send.amount', 'Amount') }}</label>
              <div class="col-sm-10">
                <div class="input-group mb-1">
                  <input class="form-control"
                         @input="formatAmount"
                         required
                         id="inputAmount"
                         :class="{':invalid':notEnoughBalance}"
                         v-model="amount">
                  <div class="input-group-append">
                    <span class="input-group-text" id="basic-addon1">{{ currency }}</span>
                  </div>
                  <button v-if="currency === 'XES'" class="btn btn-link ml-2" type="button" id="button-addon2" @click="max">{{ $t('max', 'Max')}}</button>
                </div>
                <div class="text-danger" v-if="notEnoughBalance">
                  <small>{{ $t('wallet.send.notEnoughBalance', 'Your {currency} balance is too low for this amount', {currency})}}</small>
                </div>
              </div>
            </div>
          </form>
        </div>
      </div>
      <div class="actions px-3 py-2 d-flex bg-light">
        <button class="btn btn-secondary ml-auto" @click="backToOverview">{{ $t('generic.button.cancel', 'Cancel') }}</button>
        <button class="btn btn-primary ml-2" type="submit" @click="next" :disabled="formValid === false">{{ $t('wallet.send.next', 'Next')}}</button>
      </div>
    </div>
    <div v-if="step === 'confirm'">
      <div class="border-bottom">
        <button class="btn btn-link py-2" @click="step = 'form'">
          <span class="mdi mdi-chevron-left md-24 d-inline-block border-right pr-1"></span> <span
          class="text-muted d-inline-block pl-1">{{ $t('todo', 'Back') }}</span>
        </button>
      </div>
      <h2 class="text-center pt-3">{{ $t('wallet.send.confirm', 'Confirm transaction')}}</h2>
      <div class="wallet-box text-left form p-3 mb-3 mx-3">
        <div class="form-group row align-items-center">
          <label for="balance1" class="col-sm-4 col-form-label">{{ $t('wallet.send.balance', 'Balance') }}</label>
          <div class="col-sm-8">
            <span id="balance1"><span class="eth eth-nr">{{ balance }}</span><span class="eth ml-1">{{ currency }}</span></span>
          </div>
        </div>
        <hr>
        <div class="form-group row align-items-center">
          <label for="balance" class="col-sm-4 col-form-label">{{ $t('wallet.send.amount', 'Amount') }}</label>
          <div class="col-sm-8"><span><span class="eth eth-nr">{{ amount }}</span><span
            class="eth ml-1">{{ currency }}</span></span></div>
        </div>
        <div class="form-group row mb-0 align-items-center">
          <label for="balance" class="col-sm-4 col-form-label">To</label>
          <div class="col-sm-8">
            <span v-if="toAddress && toAddress.name && toAddress.name != toAddress.address">
              <strong class="address-name">{{ toAddress.name }}</strong>
              <span class="text-muted">
                <small>{{ toAddress.address }}</small>
              </span>
            </span>
            <span v-else class="text-muted"><small>{{ toAddress.address }}</small></span>
          </div>
        </div>
      </div>
      <costs v-if="gasEstimate !== undefined"
             class="mx-3 mb-3"
             :gasEstimate="gasEstimate"/>
      <div class="actions px-3 py-2 d-flex bg-light">
        <button class="btn btn-secondary ml-auto" @click="step = 'form'">{{ $t('generic.button.cancel', 'Cancel') }}</button>
        <button class="btn btn-primary ml-2" type="submit" @click="confirm" :disabled="formValid === false || sending">
          {{ $t('confirm', 'Confirm') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import Multiselect from 'vue-multiselect'
import { isAddress } from 'web3-utils'
import Costs from '@/components/Costs'

export default {
  name: 'send',
  data () {
    return {
      currency: 'XES',
      toAddress: '',
      tags: [],
      step: 'form',
      amount: '',
      sending: false,
      gasEstimate: undefined
    }
  },
  components: {
    Multiselect,
    Costs
  },
  computed: {
    contacts () {
      return this.tags.concat(this.$store.getters.addressesWithoutMyself)
    },
    balance () {
      if (this.currency === 'XES') {
        return this.$store.state.wallet.balance
      }
      return this.$store.state.wallet.ethBalance
    },
    formValid () {
      return parseFloat(this.amount) <= this.balance && this.amount.toString().length > 0 && this.amount > 0 && this.toAddress !== '' && isAddress(this.toAddress.address) === true
    },
    notEnoughBalance () {
      return parseFloat(this.balance) < parseFloat(this.amount)
    },
    addressEmptyOrValid () {
      return !this.toAddress || (this.toAddress.address !== '' && isAddress(this.toAddress.address) === true)
    }
  },
  watch: {
    currency: function () {
      this.amount = ''
    },
    step () {
      if (this.step === 'confirm') {
        this.estimateGas()
      } else {
        this.gasEstimate = undefined
      }
    }
  },
  methods: {
    close () {
      this.$emit('modalClosed')
    },
    formatAmount (event) {
      this.amount = this.amount.toString().replace(/[^0-9.]+/g, '')
    },
    max () {
      this.amount = this.balance
    },
    next () {
      this.step = 'confirm'
    },
    backToOverview () {
      this.$store.commit('UPDATE_ACTIVE_WALLET_PANE', 'overview')
    },
    setCustomAddress (address) {
      const addr = {
        name: address,
        address: address
      }
      this.toAddress = addr
      this.tags.push(addr)
    },
    async estimateGas () {
      let response = null
      switch (this.currency) {
        case 'XES':
          response = await this.$store.dispatch('SEND_XES_ESTIMATE_GAS', { amount: this.amount, address: this.toAddress.address })
          break
        case 'ETH':
          response = await this.$store.dispatch('SEND_ETH_ESTIMATE_GAS', { amount: this.amount, address: this.toAddress.address })
          break
      }

      if (response && response.data) {
        this.gasEstimate = {
          gasPrice: response.data.gasPrice,
          gasLimit: response.data.gasLimit
        }
      } else {
        this.gasEstimate = false
      }
    },
    async confirm () {
      this.sending = true
      let res = null

      switch (this.currency) {
        case 'XES':
          res = await this.$store.dispatch('SEND_XES', { amount: this.amount, address: this.toAddress.address })
          break
        case 'ETH':
          res = await this.$store.dispatch('SEND_ETH', { amount: this.amount, address: this.toAddress.address })
          break
        default:
          break
      }

      if (res && res.status === true) {
        this.$notify({
          title: this.$t('wallet.send.success', 'Sending {amount} {currency} to {address}', {
            amount: this.amount,
            currency: this.currency,
            address: this.toAddress.name || this.toAddress.address
          }),
          type: 'success'
        })
        this.backToOverview()
        this.sending = false
        this.close()
      } else {
        this.$notify({
          title: this.$t('wallet.send.error', 'Transaction error'),
          type: 'error'
        })
        this.sending = false
      }
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .wallet-box {
    border-radius: $border-radius;
    background: $light;

    .address-name {
      word-wrap: break-word;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      max-width: 320px;
      display: block;
    }
  }

  .xes,
  .eth {
    color: $primary;
    font-size: 1.5rem;
    font-weight: 100;

    &.xes-nr,
    &.eth-nr {
      font-weight: 400;
    }
  }

  .actions {
    border-bottom-right-radius: $border-radius;
    border-bottom-left-radius: $border-radius;
  }

  .icon-transaction-successs {
    font-size: 4rem;
  }

  /deep/ .multiselect__tags .multiselect__single {
    overflow: hidden;
  }
</style>
