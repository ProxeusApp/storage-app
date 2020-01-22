<template>
  <div class="langSelectWrapper">
    <form>
      <label for="langSelect" class="pr-1">{{ $t('filebrowser.buttons.language', 'Language') }}</label>
      <select name="langSelect" id="langSelect" v-model="selectedLanguage">
        <option v-for="lang in availableLanguages" :key="lang" :value="lang">{{ lang }}</option>
      </select>
    </form>
  </div>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'language-select',

  computed: {
    selectedLanguage: {
      get () {
        return this.language
      },
      set (lang) {
        this.changeLanguage(lang)
      }
    },
    ...mapState({
      availableLanguages: state => state.availableLanguages,
      language: state => state.language
    })
  },
  methods: {
    changeLanguage (newLanguage) {
      this.$store.commit('CHANGE_LANGUAGE', newLanguage)
      this.$i18n.set(newLanguage)
    }
  }
}
</script>

<style lang="scss" scoped>
  @import "../assets/styles/variables.scss";
</style>
