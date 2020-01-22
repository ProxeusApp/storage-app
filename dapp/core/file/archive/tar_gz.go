package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProxeusApp/pgp"
)

type ProxeusMeta struct {
	Version     int
	Kind        string
	FileNameMap map[string]string
	ProcessName string
}

const (
	proxeusTarGzV1FileNameSig = "00000_e166801d00a45901e2b3ca692a6a95e367d4a976218b485546a2da464b6c88b5"
	proxeusTarGzV2FileNameSig = "00001_682203e1f882ff4fad65a0a72abee663558948853b437f21dfdb7e38a16eb366"
	proxeusTarGzV3FileNameSig = "00002_ks32yml3lsj2xf4fad65a0a72abme66359slwmko3b437f21dfdb7123123wpa6"
	KindProxeusProcess        = "process"

	Thumb          = "thumb"
	ThumbEncrypted = "thumb_encrypted"
)

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)w
func TarFileList(proxMetaArmoured []byte, srcList []string, writer io.Writer) error {
	gzw := gzip.NewWriter(writer)

	tw := tar.NewWriter(gzw)

	err := addProxeusMeta(proxMetaArmoured, tw)
	if err != nil {
		log.Println("TarFileList addProxeusMeta error ", err)
		return err
	}

	for _, file := range srcList {
		fi, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("Unable to tar files - %v", err.Error())
		}
		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			log.Println("TarFileList FileInfoHeader error ", err)
			return err
		}

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			log.Println("TarFileList WriteHeader error ", err)
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			log.Println("TarFileList open file error ", err)
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			log.Println("TarFileList file copy error ", err)
			return err
		}
		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()
	}
	if err != nil {
		log.Println("TarFileList file error ", err)
		return err
	}
	err = tw.Close()
	if err != nil {
		log.Println("TarFileList tarwriter close error ", err)
		return err
	}
	err = gzw.Close()
	if err != nil {
		log.Println("TarFileList gzipwriter close error ", err)
	}
	return err
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)w
func Tar(src string, writer io.Writer) error {
	var err error
	// ensure the src actually exists before trying to tar it
	if _, err = os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	gzw := gzip.NewWriter(writer)

	tw := tar.NewWriter(gzw)

	// walk path
	err = filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
	if err != nil {
		return err
	}
	err = tw.Close()
	if err != nil {
		return err
	}
	err = gzw.Close()
	return err
}

var ErrNoProxeusArchive = errors.New("not a Proxeus archive")
var ErrWhenParsingProxeusMeta = errors.New("error when parsing Proxeus meta")

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func UntarProxeusArchive(dst string, r io.Reader, pw, pgpPriv []byte) (pm *ProxeusMeta, err error) {
	_, err = os.Stat(dst)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dst, 0750)
		if err != nil {
			return
		}
	}
	var gzr *gzip.Reader
	gzr, err = gzip.NewReader(r)
	if err != nil {
		return
	}
	isProxeusArchive := false
	defer gzr.Close()
	defer func() {
		if !isProxeusArchive {
			log.Println("[archive][UntarProxeusArchive] error:", err)
			err = ErrNoProxeusArchive
			os.RemoveAll(dst)
		}
	}()

	tr := tar.NewReader(gzr)
	pm = &ProxeusMeta{}
	decryptFlag := false
	for {
		decryptFlag = false
		header, er := tr.Next()

		switch {
		// if no more files are found return
		case er == io.EOF:
			return
			// return any other error
		case er != nil:
			return
			// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		var target string
		if header.Name == proxeusTarGzV1FileNameSig ||
			header.Name == proxeusTarGzV2FileNameSig {
			isProxeusArchive = true
			var bts []byte
			bts, err = ioutil.ReadAll(tr)
			if err != nil {
				err = ErrWhenParsingProxeusMeta
				isProxeusArchive = false
				return
			}
			err = json.Unmarshal(bts, pm)
			if err != nil {
				err = ErrWhenParsingProxeusMeta
				isProxeusArchive = false
				return
			}
			continue
		} else if header.Name == proxeusTarGzV3FileNameSig {
			isProxeusArchive = true
			decryptFlag = true

			bts := bytes.NewBuffer(nil)

			_, err = pgp.DecryptStream(tr, bts, pw, pgpPriv)
			if err != nil {
				log.Println("[archive][UntarProxeusArchive] error while decrypting stream:", err)
				return
			}

			err = json.Unmarshal(bts.Bytes(), pm)
			if err != nil {
				log.Println("[archive][UntarProxeusArchive] error while unmarshal:", err)
				err = ErrWhenParsingProxeusMeta
				isProxeusArchive = false
				return
			}
			continue
		} else if header.Name == ThumbEncrypted {
			decryptFlag = true
			target = filepath.Join(dst, Thumb)
		} else {
			target = filepath.Join(dst, header.Name)
			// the target location where the dir/file should be created
			if len(pm.FileNameMap) > 0 {
				//if header.Name is actualName we are dealing with the file so we decrypt, else its the plain thumb
				if actualName, ok := pm.FileNameMap[header.Name]; ok {
					if actualName != "" {
						target = filepath.Join(dst, actualName)
						decryptFlag = true
					}
				}
			}
		}

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err = os.Stat(target); err != nil {
				if err = os.MkdirAll(target, 0755); err != nil {
					return
				}
			}
			// if it's a file create it
		case tar.TypeReg:
			var f *os.File
			f, err = os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return
			}
			if decryptFlag && len(pgpPriv) > 0 {
				_, err = pgp.DecryptStream(tr, f, pw, pgpPriv)
				if err != nil {
					f.Close()
					return
				}
			} else {
				// copy over contents
				if _, err = io.Copy(f, tr); err != nil {
					f.Close()
					return
				}
			}
			f.Close()
		}
	}
}

func Untar(dst string, r io.Reader) (err error) {
	_, err = os.Stat(dst)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dst, 0750)
		if err != nil {
			return
		}
	}
	var gzr *gzip.Reader
	gzr, err = gzip.NewReader(r)
	if err != nil {
		return
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, er := tr.Next()

		switch {
		// if no more files are found return
		case er == io.EOF:
			return
			// return any other error
		case er != nil:
			return
			// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)
		// the target location where the dir/file should be created

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err = os.Stat(target); err != nil {
				if err = os.MkdirAll(target, 0755); err != nil {
					return
				}
			}
			// if it's a file create it
		case tar.TypeReg:
			var f *os.File
			f, err = os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return
			}
			// copy over contents
			if _, err = io.Copy(f, tr); err != nil {
				f.Close()
				return
			}
			f.Close()
		}
	}
}

func addProxeusMeta(proxMetaArmoured []byte, tw *tar.Writer) error {
	if proxMetaArmoured == nil {
		return nil
	}
	hdr := &tar.Header{
		Name:     proxeusTarGzV3FileNameSig,
		Mode:     0600,
		Size:     int64(len(proxMetaArmoured)),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(proxMetaArmoured)
	return err
}
