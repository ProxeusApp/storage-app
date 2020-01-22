document.addEventListener('astilectron-ready', () => {
  const appName = 'Proxeus Storage'

  if (typeof electron === 'object') {

    //PSS-87: clear electron cache on window close (handle in frontend, did not work with astilectron)
    const electronWindow = electron.remote.getCurrentWindow();
    window.addEventListener('beforeunload',  () => {
      electronWindow.webContents.session.clearCache(()=> {
        console.log("Electron Cache has been cleared") //clearCache only works reliably when callback defined
      });
    });

    const template = [
      {
        label: 'Edit',
        submenu: [
          { role: 'undo' },
          { role: 'redo' },
          { type: 'separator' },
          { role: 'cut' },
          { role: 'copy' },
          { role: 'paste' },
          { role: 'pasteandmatchstyle' },
          { role: 'delete' },
          { role: 'selectall' }
        ]
      },
      {
        role: 'window',
        submenu: [
          { role: 'minimize' },
          { role: 'close' }
        ]
      }
    ]

    // osPlatform is set in main.go#167
    if (osPlatform === 'darwin') {
      template.unshift({
        label: appName,
        submenu: [
          { role: 'hide', label: `Hide ${appName}` },
          { role: 'hideothers' },
          { role: 'unhide' },
          { type: 'separator' },
          { role: 'quit', label: `Quit ${appName}` }
        ]
      })
    }

    let menu = electron.remote.Menu.buildFromTemplate(template)
    electron.remote.Menu.setApplicationMenu(menu)
  }
})

window.openInBrowser = (event, element) => {
  // if in electron app context open in new electron window
  // else we are in browser context and links are opened normally
  if (typeof electron === 'object') {
    event.preventDefault()
    electron.remote.shell.openExternal(element.getAttribute('href'))
    return false
  }
}
