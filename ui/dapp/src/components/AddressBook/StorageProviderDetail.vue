<template>
  <div class="detail-view w-100 align-items-center text-left">
    <div class="back d-flex">
      <button class="btn-back btn btn-link py-0 px-2 w-100 border-bottom" @click="backToList"><i class="mdi mdi-chevron-left mdi-24px"></i> {{ $t('addressbook.storageProvider.backToList', 'Back to list') }}</button>
    </div>
    <div class="scroll d-flex flex-column">
      <div class="px-2">
        <div v-if="storageProvider.logoUrl" class="logo text-center m-3">
          <img :src="storageProvider.logoUrl" :alt="storageProvider.name + ' Logo'" height="70">
        </div>
        <h2 class="text-center m-3">{{ storageProvider.name }}</h2>
        <table class="table">
          <tr v-if="storageProvider.description">
            <td>{{ $t('addressbook.storageProvider.description', 'Description') }}</td>
            <td>{{ storageProvider.description }}</td>
          </tr>
          <tr v-if="storageProvider.address">
            <td>{{ $t('addressbook.storageProvider.address', 'Address') }}</td>
            <td class="text-break">{{ storageProvider.address }}</td>
          </tr>
          <tr>
            <td>{{ $t('addressbook.storageProvider.contract', 'Contract') }}</td>
            <td>
              <a :href="contractUrl" target="_blank" onclick="window.openInBrowser(event, this);">
                {{ $t('addressbook.storageProvider.contractOnEtherscan', 'Contract on Etherscan') }} <i class="mdi mdi-launch"></i>
              </a>
            </td>
          </tr>
          <tr v-if="storageProvider.jurisdictionCountry">
            <td>{{ $t('addressbook.storageProvider.jurisdictionCountry', 'Jurisdiction') }}</td>
            <td>{{ storageProvider.jurisdictionCountry }}</td>
          </tr>
          <tr v-if="storageProvider.dataCenter">
            <td>{{ $t('addressbook.storageProvider.dataCenter', 'Data center') }}</td>
            <td>{{ storageProvider.dataCenter }}</td>
          </tr>
          <tr v-if="storageProvider.termsAndConditionsUrl || storageProvider.privacyPolicyUrl">
            <td>{{ $t('addressbook.storageProvider.documents', 'Documents') }}</td>
            <td>
              <template v-if="storageProvider.termsAndConditionsUrl">
                <a :href="storageProvider.termsAndConditionsUrl" target="_blank" onclick="window.openInBrowser(event, this);">{{ $t('addressbook.storageProvider.termsAndConditions', 'Terms and conditions') }}</a><br>
              </template>
              <a v-if="storageProvider.privacyPolicyUrl" :href="storageProvider.privacyPolicyUrl" target="_blank" onclick="window.openInBrowser(event, this);">{{ $t('addressbook.storageProvider.privacyPolicy', 'Privacy policy') }}</a>
            </td>
          </tr>
          <tr v-if="maxFileSize">
            <td>{{ $t('addressbook.storageProvider.maxFileSize', 'Max File Size') }}</td>
            <td>{{ maxFileSize }} MB</td>
          </tr>
          <tr v-if="maxStorageDuration">
            <td>{{ $t('addressbook.storageProvider.maxStorageDuration', 'Max Storage Duration') }}</td>
            <td>{{ maxStorageDuration }} {{ $t('addressbook.storageProvider.days', 'Days') }}</td>
          </tr>
          <tr v-if="pricePerKB">
            <td>{{ $t('addressbook.storageProvider.pricePerKB', 'Storage price per KB') }}</td>
            <td>{{ pricePerKB }} XES</td>
          </tr>
          <tr v-if="pricePerDay">
            <td>{{ $t('addressbook.storageProvider.pricePerDay', 'Storage price per day') }}</td>
            <td>{{ pricePerDay }} XES</td>
          </tr>
          <tr v-if="gracePeriod">
            <td>{{ $t('addressbook.storageProvider.gracePeriod', 'Grace period') }}</td>
            <td>{{ gracePeriod }}</td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import moment from 'moment'

export default {
  name: 'storage-provider-detail',
  computed: {
    storageProvider () {
      return this.$store.state.address.activeStorageProviderDetail
    },
    maxFileSize () {
      return (this.storageProvider.maxFileSizeByte / 1024 / 1024).toFixed(1)
    },
    pricePerKB () {
      return this.storageProvider.priceByte * 1024
    },
    pricePerDay () {
      return this.storageProvider.priceDay
    },
    maxStorageDuration () {
      return this.storageProvider.maxStorageDays
    },
    gracePeriod () {
      return moment.duration(this.storageProvider.graceSeconds, 'seconds').humanize()
    },
    contractUrl () {
      const etherscanUrl = this.$store.getters.etherscanUrl
      return `${etherscanUrl}/address/${this.storageProvider.address}`
    }
  },
  methods: {
    backToList () {
      this.$store.commit('SET_STORAGE_PROVIDER_VIEW', { view: 'list' })
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .table {
    font-size: $font-size-sm;
    max-width: 100%;

    td:nth-child(2) {
      //word-break: break-all;
    }

    tr > td:first-child {
      font-weight: 500;
    }
  }

  .detail-view {
    position: relative;
  }

  .scroll {
    height: calc(100vh - 50px - 55px - 95px);
    padding-top: 50px;
    overflow-y: auto;
  }

  .btn-back {
    background: $gray-100;
    height: 50px;
    width: 100%;
    position: absolute;
    text-align: left;

    &:hover {
      background: $gray-200;
    }
  }
</style>
