class FileSystem {
  constructor () {
    this.initialized = false
    this.fileSystem = undefined
    this.initialize()
  }

  initialize () {
    return new Promise((resolve, reject) => {
      if (this.initialized) {
        resolve()
      }
      window.requestFileSystem = window.requestFileSystem ||
        window.webkitRequestFileSystem
      window.requestFileSystem(window.PERMANENT, 1024 * 1024 * 100, (fs) => {
        this.fileSystem = fs
        this.initialized = true
        resolve()
      }, (e) => {
        reject(e)
      })
    })
  }

  createFile (file, hash) {
    return new Promise((resolve, reject) => {
      this.fileSystem.root.getFile(hash, { create: true }, (fileEntry) => {
        fileEntry.createWriter((fileWriter) => {
          fileWriter.onwriteend = () => {
            console.log('Write completed.')
            resolve()
          }
          fileWriter.onerror = (e) => {
            console.log('Write failed: ' + e.toString())
            reject(e)
          }
          fileWriter.write(file)
        }, (e) => {
          reject(e)
        })
      }, (e) => {
        reject(e)
      })
    })
  }

  readFile (fileName) {
    return new Promise((resolve, reject) => {
      this.fileSystem.root.getFile(fileName, { create: false }, (fileEntry) => {
        resolve(fileEntry)
      }, (e) => {
        console.log(e)
        reject(e)
      })
    })
  }

  readDirectory () {
    return new Promise((resolve, reject) => {
      const dirReader = this.fileSystem.root.createReader()
      dirReader.readEntries((results) => {
        resolve(results)
      }, (e) => {
        console.log(e)
        reject(e)
      })
    })
  }

  removeFile (fileName) {
    return new Promise((resolve, reject) => {
      this.fileSystem.root.getFile(fileName, { create: false }, (fileEntry) => {
        fileEntry.remove(() => {
          console.log('File removed.')
          resolve()
        }, (e) => {
          reject(e)
        })
      }, (e) => {
        reject(e)
      })
    })
  }
}

const fs = new FileSystem()
fs.initialize()
export default fs
