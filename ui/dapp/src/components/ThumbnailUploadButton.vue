<template>
  <form enctype="multipart/form-data" novalidate>
    <div class="mdi mdi-pencil edit-icon">
      <input class="btn-file-input" type="file" name="file" accept="image/*" :disabled="costEstimationInProgress"
             @change="filesChange($event.target.files);"/>
    </div>
  </form>
</template>

<script>
export default {
  name: 'thumbnail-upload-button',
  props: ['costEstimationInProgress'],
  data () {
    return {
      newFileThumbnail: undefined
    }
  },
  methods: {
    filesChange (fileList) {
      this.$emit('thumbnailChanged', fileList[0])
    }
  }
}
</script>

<style lang="scss" scoped>
  .btn.btn-primary {
    position: relative;
  }

  .btn-file-input {
    cursor: pointer;
    position: absolute;
    font-size: 50px;
    opacity: 0;
    right: 0;
    top: 0;
    width: 100%;
    height: 100%;
  }

  .edit-icon {
    color: white;
    text-shadow: 1px 1px 10px rgba(0, 0, 0, 0.8);
    font-size: 2.5rem;
    position: absolute;
    padding-top: calc(100% - 45px);
    padding-left: calc(100% - 45px);
    padding-bottom: 45px;
    padding-right: 45px;
    cursor: pointer;
    height: 0;
    width: 0;

    &:hover {
      color: #dddddd;
    }
  }
</style>
