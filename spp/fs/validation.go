package fs

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.proxeus.com/core/central/dapp/core/file/crypt"

	"git.proxeus.com/core/central/dapp/core/file/archive"
)

//verify all files of the archive are formatted as pgp-encrypted-file
func verifyArchivePgpFiles(src, dst string) error {
	if err := os.MkdirAll(dst, 0700); err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(dst); err != nil {
			log.Printf("[Validation][verifyArchivePgpFiles] failed to remove decompressed file, err: %s",
				err.Error())
		}
	}()

	tarFile, err := os.Open(src)
	defer tarFile.Close()
	if err = archive.Untar(dst, tarFile); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(dst)
	for _, f := range files {
		fPath := filepath.Join(dst, f.Name())
		if err = verifyPgpFormat(fPath); err != nil {
			log.Printf("[Validation][verifyArchivePgpFiles] verifyPgpFormat failed for: %s, err: %s",
				fPath, err.Error())
			return err
		}
	}

	return nil
}

var ErrFileNotPGPEncrypted = errors.New("file not pgp encrypted")

//decrypt with random private key to check for file formatting errors returned by openpgp-library
func verifyPgpFormat(filePath string) error {
	pgpFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	fileBytes, err := ioutil.ReadAll(pgpFile)
	fileContent := string(fileBytes)

	fContents := strings.Split(fileContent, "-----BEGIN PGP MESSAGE-----")
	if len(fContents) != 2 || fContents[0] != "" || fContents[1] == "" {
		return ErrFileNotPGPEncrypted
	}

	fContents = strings.Split(fContents[1], "-----END PGP MESSAGE-----")
	if len(fContents) != 2 || fContents[0] == "" || fContents[1] != "" {
		return ErrFileNotPGPEncrypted
	}

	err = crypt.DecryptFile(filePath+"_plain", filePath, []byte(""), randomValidationKey())
	defer func() {
		_ = os.Remove(filePath + "_plain")
		_ = pgpFile.Close()
	}()
	if err == nil {
		log.Fatal("[validation][verifyPgpFormat] THIS SHOULD NEVER HAPPEN. No error on decrypt with random key")
		return ErrFileNotPGPEncrypted
	}

	if err.Error() != "openpgp: incorrect key" {
		return err
	}

	return nil
}

//returns a random private key to test decryption
func randomValidationKey() []byte {
	return []byte(`-----BEGIN PGP PRIVATE KEY BLOCK-----

lQOsBFxK+0gBCACFuqxj9y4dL2BYtfN6MYBA5Fas1EXFc+xo8xXt82vgdsQE8Tah
THcMbjX/Joa/JDx14HFz+bd7SjzVbYroWW6+aks3CzDWWVS7ND8gMrsPQ6K3BH5+
gtiOBqTCLL4YPtpo903FoH7G7ZMKiO1NqpE2KXIGQlqa+E+c9ypo+l5ky6dgBRC1
xVGdKjTg2HuTTXSFWIbHu7Y1vKYjQA5fucFdMJONKhYwLklQClvHpz1M1VB6/5SK
hcCopWh02qlKF9SBIrHYNe0XurVCoa7CmieNswIku8HOTRUReriBVzDgfLfqU88N
UpY1oFXZGvIo2eTGqICv2x17alPLFBcy/fjPABEBAAH/AwMC3hzQ2cAwnQZgh22d
JweuI2t6GTjEijF4Yg9v9n7dS3pKip5p5PVhY0m8SKwgcsSHiIBgMU6Bb6Tfv1gf
H4sZspb9/O3tBBQzeDdlQ0GoDqUPgi41NCpFRtSATdUMkOXoVNaU6Oej+kHgz1Xr
HItVpmVwMzqZQpcwfUv0KbW/jOoXaQOxRf5Fp8lhAlRS96mZuy6qtkGBTKY5/QvT
nRE894r6Qhh2TvwMUpYo+KkZqwfiimtYzOdDgpd/U5e9Lrus1lFPg3WKfYhltnNh
S6GQJ0lbJIXI2AHjfROfav/NT60MrM6iLNli5C8751PdjAnBqxqvUXPTjIE5kEZK
ZT1UO9mTzrmKw+wDmiuBI7iP0vn1lHXxnGuDbPtha486gOKsKIpo5gtxWSvVOlg+
WoGMlaY1i4kqg7KbScLtJlu33CPpEwzY078xWjDD+EIqaRnlu81hWJIs8rK6/YPC
M0IvwBveCB9kD3C9IKiI8jHjoX0C0RFARobOeH9aR3ywn+2Uf4soviq6VbN7gtcK
t+jeD2ssa75jpxVfgvHEolONTwnzwFD+FieDYULuoK0CC+PNaXNhGHIZUTOw8+dF
Z5C/l+PgdsFPWSnu6+cim1fAhP6wys5+mdE84qz8XGzrsB/qk3ZAKb4yC1Eyiuky
EqFkcpP+ysOYG6a50KAkEc4P1ugCkV1NRHDjfZ7pDU73B2qUpTCjcXAzhX9eowHg
C1wbw+mrXtGWotP6FqTYjKDkq6Pn90piIDkCIcxEuvNiFhYxSmkfPHABQhiFZXLT
aQTCdUKc/LOrIdJszKP5z+HIpBOw3Wlaclcy4llpMoaeQxDPUk/y8KPzXiH39RL6
sEMi9sDvFGnGt8CskGnFYkVTaaK+dFnQKfQ5Nk8DebQRYXNkYXNkQGFzZGFzZC5j
b22JARwEEAECAAYFAlxK+0gACgkQdGwoN7SWZnMmwAf/dHL1ag4lyW1KP4E3vwuV
ww4sCjL2LztmVGYgSEgl3MfqSl10J3BgAnD14fSOMNv4kFnJS2UHfeX7inmE4YVL
FHe7unIaXc4uUGWr0fvVPP+fJwsk/SjdIg8Wv5dBcvtsncFPzI66c/sqvc9qm1qc
jGyGGB46xnJ96+hNvYDHvkgJK2iwNDqOYxEZZcokNe2YHHHrgJZXhB3+8O8WT8il
Xue50L8aO2Pq5217oEzrrhn1Y8wrI1CwesG9wyWI9Hhmh+uEkiNg4oQSwDR0FrZr
I0fQd2t64CBwzRK+0d+iBPt3zHjZQ50RIkjiecNLeLXvAO00EwKgnFHRYPODC67U
Ag==
=xX9P
-----END PGP PRIVATE KEY BLOCK-----
`)
}
