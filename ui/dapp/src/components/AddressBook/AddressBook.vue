<template>
  <div class="address-book text-center">
    <b-tabs class="nav-justified w-100">
      <b-tab :title="$t('addressbook.tabs.contacts', 'Contacts')" active id="contacts-tab" class="tab-content p-3">
        <button class="btn btn-primary btn-round mb-3" @click="toggleNewAddressVisible">
          <span class="mdi md-18 mdi-plus"></span>
        </button>

        <search-box :placeholder="$t('addressbook.placeholder.search','Search addresses…')"
                    :term="searchTerm"
                    class="mb-2"
                    @search="search"></search-box>

        <transition name="scale-show">
          <div v-if="newAddressVisible" class="new-address-wrapper mb-2">
            <div class="new-address text-left">
              <form @submit.prevent="addAddress" novalidate>
                <div class="form-group">
                  <!--<label for="exampleInputEmail1">Email address</label>-->
                  <input type="text" ref="inputNewName" class="form-control" id="inputNewName"
                         aria-describedby="emailHelp"
                         v-model.trim="newName"
                         :placeholder="$t('addressbook.placeholder.name', 'Name')"
                         :class="{'is-invalid':!isNameValid && wasValidated}">
                  <div class="invalid-feedback">
                    {{ $t('addressbook.validation.nameRequired', 'A name is required') }}
                  </div>
                </div>
                <div class="form-group">
                  <!--<label for="exampleInputEmail1">Email address</label>-->
                  <input type="text" class="form-control" id="inputNewPk" aria-describedby="emailHelp"
                         v-model.trim="newAddress"
                         :placeholder="$t('addressbook.placeholder.ethereumAccount', 'Ethereum Account Address')"
                         :class="{'is-invalid':!isAddressValid && wasValidated}">
                  <div class="invalid-feedback">
                    {{ $t('addressbook.validation.etherAddressInvalid', 'Not a valid ethereum address') }}
                  </div>
                </div>
                <button :disabled="!areAllFieldsValid" type="submit" class="btn btn-sm btn-primary mr-1">{{ $t('addressbook.contacts.addaddress', 'Add address') }}</button>
                <button type="button" class="btn btn-sm btn-secondary" @click="toggleNewAddressVisible">{{ $t('generic.button.cancel', 'Cancel') }}</button>
              </form>
            </div>
          </div>
        </transition>

        <div class="address-list text-left w-100" :class="{'new-form-visible':newAddressVisible}">
          <address-book-entry v-for="address in addresses"
                              :key="address.address"
                              :address="address"
                              @removed="removeAddressBookEntry"
                              @changedName="updateAddressBookEntryName"/>
        </div>
      </b-tab>
      <b-tab :title="$t('addressbook.tabs.providers', 'Storage Providers')" id="storage-providers-tab"
             class="tab-content">
        <div class="storage-provider-view">
          <transition :name="storageProviderViewTransitionName">
            <component :is="storageProviderView"></component>
          </transition>
        </div>
      </b-tab>
    </b-tabs>
  </div>
</template>

<script>
import AddressBookEntry from './AddressBookEntry'
import StorageProviderList from './StorageProviderList'
import StorageProviderDetail from './StorageProviderDetail'
import SearchBox from '../SearchBox'
import web3Utils from 'web3-utils'

export default {
  name: 'address-book',
  components: {
    AddressBookEntry,
    SearchBox,
    StorageProviderList,
    StorageProviderDetail
  },
  data () {
    return {
      searchTerm: '',
      newAddressVisible: false,
      newName: '',
      newAddress: '',
      wasValidated: false
    }
  },
  computed: {
    storageProviderViewTransitionName () {
      return this.storageProviderView === 'StorageProviderList' ? 'prev' : 'next'
    },
    storageProviderView () {
      return this.$store.state.address.storageProviderView === 'list' ? 'StorageProviderList' : 'StorageProviderDetail'
    },
    myAddress () {
      return this.$store.getters.myself
    },
    addresses () {
      return this.$store.getters.addressesBySearchTerm(this.searchTerm)
    },
    storageProviders () {
      return this.$store.getters.activeStorageProviders
    },
    isNameValid () {
      return this.newName !== ''
    },
    isAddressValid () {
      return this.newAddress !== '' && this.validAddress(this.newAddress)
    },
    areAllFieldsValid () {
      return this.isNameValid && this.isAddressValid && this.wasValidated
    }
  },
  methods: {
    search (term) {
      this.searchTerm = term
    },
    toggleNewAddressVisible () {
      this.newAddressVisible = !this.newAddressVisible
      this.newAddressVisible === true && this.$nextTick(() => {
        this.$refs.inputNewName.focus()
      })
    },
    addAddress () {
      this.$store.dispatch('ADD_ADDRESS',
        { name: this.newName, address: this.newAddress })
      this.newName = ''
      this.newAddress = ''
      // this.pgpPublicKey = undefined
      this.wasValidated = false
      this.newAddressVisible = false
    },
    removeAddressBookEntry (address) {
      this.$store.commit('SET_ADDRESS_TO_REMOVE', address)
      if (this.$store.state.address.doNotShowContactRemoveWarning === true) {
        this.$store.dispatch('REMOVE_ADDRESS', this.$store.state.address.addressToRemove)
        this.$store.commit('SET_REMOVE_CONTACT_WARNING_MODAL', false)
      } else {
        this.$store.commit('SET_REMOVE_CONTACT_WARNING_MODAL', true)
        this.$emit('closeAddressBook')
      }
    },
    updateAddressBookEntryName ({ address, name }) {
      this.$store.dispatch('UPDATE_ADDRESS_NAME', { address, name })
    },
    validAddress (address) {
      this.wasValidated = true
      return web3Utils.isAddress(address)
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../../assets/styles/variables";

  .textarea-sm {
    resize: none;
    font-size: 12px;
  }

  .scale-show-enter-active,
  .scale-show-leave-active {
    transition: transform 200ms ease, max-height 200ms ease;
  }

  .scale-show-enter,
  .scale-show-leave-to {
    transform: scale(0);
    max-height: 0;
  }

  .scale-show-enter-to,
  .scale-show-leave {
    transform: scale(1);
    max-height: 300px;
  }

  .address-book {
    background: $gray-50;
    height: 100%;
    border-radius: $border-radius-lg;
    overflow: hidden;
  }

  .new-address-wrapper {
    overflow: hidden;
  }

  .new-address {
    width: 100%;
    border-radius: $border-radius;
    padding: $spacer;
    background-color: $gray-200;
  }

  /deep/ .tab-content {
    position: relative;
    min-height: calc(100vh - 150px);
    max-height: calc(100vh - 150px);
    overflow: auto;
  }

  /deep/ .nav-tabs {
    .nav-link {
      color: white;
      background: $primary;
      border: none;
      padding: 0.85rem 0.5rem;

      &.active {
        color: $primary;
        background: transparent;
        border: none;
      }
    }
  }

  .storage-provider-view {
    min-height: 100%;
    display: grid;
    grid-template: "main";
    flex: 1;
    position: relative;
    z-index: 0;
    overflow-x: hidden;
    background: $gray-50;
  }

  .storage-provider-view > * {
    grid-area: main; /* Transition: make sections overlap on same cell */
    position: relative;
    background: $gray-50;
  }

  .storage-provider-view > :first-child {
    z-index: 1; /* Prevent flickering on first frame when transition classes not added yet */
  }

  /* Transitions */

  .next-leave-to {
    animation: leaveToLeft 300ms both cubic-bezier(0.165, 0.84, 0.44, 1);
    z-index: 0;
  }

  .next-enter-to {
    animation: enterFromRight 300ms both cubic-bezier(0.165, 0.84, 0.44, 1);
    z-index: 1;
  }

  .prev-leave-to {
    animation: leaveToRight 300ms both cubic-bezier(0.165, 0.84, 0.44, 1);
    z-index: 1;
  }

  .prev-enter-to {
    animation: enterFromLeft 300ms both cubic-bezier(0.165, 0.84, 0.44, 1);
    z-index: 0;
  }

  @keyframes leaveToLeft {
    from {
      transform: translateX(0);
    }

    to {
      transform: translateX(-100%);
      /*opacity: 0;*/
    }
  }

  @keyframes enterFromLeft {
    from {
      transform: translateX(-100%);
      /*opacity: 0;*/
    }

    to {
      transform: translateX(0);
    }
  }

  @keyframes leaveToRight {
    from {
      transform: translateX(0);
      /*opacity: 1;*/
    }

    to {
      transform: translateX(100%);
      /*opacity: 0;*/
    }
  }

  @keyframes enterFromRight {
    from {
      transform: translateX(100%);
      /*opacity: 0;*/
    }

    to {
      transform: translateX(0);
      /*opacity: 1;*/
    }
  }
</style>
