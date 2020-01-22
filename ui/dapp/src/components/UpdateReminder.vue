<template>
  <div class="update-reminder mx-auto mt-5">
    <!--<language-select/>-->
    <h2 class="mt-4">{{ $t('general.important.update', 'Important Update') }}</h2>
    <p v-html="text"></p>
  </div>
</template>

<script>
// import LanguageSelect from './LanguageSelect'

export default {
  name: 'update-reminder',
  components: {
    // LanguageSelect
  },
  data () {
    return {
      text: this.$t('general.new_version.available',
        'A new version of the Proxeus DApp is available.<br>Please download it at <a onclick="window.openInBrowser(event, this);" href="https://drive.google.com/drive/u/0/folders/1i9qWcx-QXedlrTMjUR2uvr2rk-h2G4kw?ths=true" target="_blank">our page</a>')
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../assets/styles/variables.scss";

  .update-reminder {
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
