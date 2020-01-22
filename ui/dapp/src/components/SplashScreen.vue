<template>
  <div class="splash-screen mx-auto mt-5">
    <!--<language-select/>-->
    <h1 class="huge mb-3">{{ $t('splashScreen.title', 'Hey there,') }}</h1>
    <p>{{ $t('splashScreen.text1', 'This is the Beta version of Samba in a test environment. Any files stored will be deleted on Beta being decommissioned.') }}</p>
    <p>{{ $t('splashScreen.text2', 'The public keys of the Ethereum and PGP Identification are logged on our PGP-Key-Server upon registration. No other account information is being collected.') }}</p>
    <!--Use v-html to show raw html defined in translation-->
    <p v-html="$t('splashScreen.text3', 'Please note that by storing any content using Tango Beta, you will enter a direct agreement with our partner and secure storage expert Mount10.')"></p>
    <button class="btn btn-primary mt-2" @click="confirmSplashScreen">{{ $t('confirm', 'Confirm') }}</button>
  </div>
</template>

<script>
// import LanguageSelect from './LanguageSelect'

export default {
  name: 'splash-screen',
  components: {
    // LanguageSelect
  },
  methods: {
    confirmSplashScreen () {
      this.$store.commit('CONFIRM_SPLASH_SCREEN')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../assets/styles/variables.scss";

  .splash-screen {
    width: 50%;
    min-width: 400px;
    max-width: 500px;

    .langSelectWrapper {
      padding-left: 0;
      border-bottom: 1px solid $gray-300;
      margin-bottom: 1rem;
      padding-bottom: 1rem;
    }
  }

  .huge {
    font-size: 3rem;
    margin-bottom: 2rem;
  }

  p {
    color: $gray-500;
    font-size: 1.1rem;
  }
</style>
