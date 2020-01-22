<template>
  <div class="search-box">
    <input type="text" class="form-control" :class="{'border-active': searchTerm ? searchTerm.length > 0 : false}"
           v-model="searchTerm" :placeholder="placeholder" ref="searchbox">
  </div>
</template>
<script>
export default {
  name: 'search-box',
  props: ['placeholder', 'term'],
  computed: {
    searchTerm: {
      get () {
        return this.term
      },
      set (value) {
        this.$emit('search', value)
      }
    }
  }
}
</script>
<style lang="scss" scoped>
  @import "../assets/styles/variables";

  .border-active {
    border: 2px solid $gray-500;
  }

  .search-box .form-control {
    height: 50px;
    border: 2px solid darken($gray-300, 1%);
  }

  .search-box .form-control:focus {
    border-color: $gray-300 !important;
  }
</style>
