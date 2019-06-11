package blsgen

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	ffi_bls "github.com/harmony-one/bls/ffi/go/bls"
	"github.com/harmony-one/harmony/crypto/bls"
)

func toISO8601(t time.Time) string {
	var tz string
	name, offset := t.Zone()
	if name == "UTC" {
		tz = "Z"
	} else {
		tz = fmt.Sprintf("%03d00", offset/3600)
	}
	return fmt.Sprintf("%04d-%02d-%02dT%02d-%02d-%02d.%09d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}

func keyFileName(publicKey *ffi_bls.PublicKey) string {
	ts := time.Now().UTC()
	serializedPublicKey := publicKey.SerializeToHexStr()
	return fmt.Sprintf("UTC--%s--bls_%s", toISO8601(ts), serializedPublicKey)
}

// GenBlsKeyWithPassPhrase generates bls key with passphrase and write into disk.
func GenBlsKeyWithPassPhrase(passphrase string) (*ffi_bls.SecretKey, string) {
	privateKey := bls.RandPrivateKey()
	publickKey := privateKey.GetPublicKey()
	fileName := keyFileName(publickKey)
	privateKeyHex := privateKey.SerializeToHexStr()
	// Encrypt with passphrase
	encryptedPrivateKeyBytes := encrypt([]byte(privateKeyHex), passphrase)
	// Write to file.
	err := ioutil.WriteFile(fileName, encryptedPrivateKeyBytes, 0600)
	check(err)
	return privateKey, fileName
}

// LoadBlsKeyWithPassPhrase loads bls key with passphrase.
func LoadBlsKeyWithPassPhrase(fileName, passphrase string) *ffi_bls.SecretKey {
	encryptedPrivateKeyBytes, err := ioutil.ReadFile(fileName)
	check(err)
	decryptedBytes := decrypt(encryptedPrivateKeyBytes, passphrase)

	priKey := &ffi_bls.SecretKey{}
	priKey.DeserializeHexStr(string(decryptedBytes))
	return priKey
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func check(err error) {
	// handle this error
	if err != nil {
		fmt.Println(err)
	}
}