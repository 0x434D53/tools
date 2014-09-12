package filecrypt

import (
	"code.google.com/p/go.crypto/blowfish"
	"code.google.com/p/go.crypto/sha3"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io/ioutil"
)

func EncryptFile(source, dst string, key []byte) error {
	sourcebytes, err := ioutil.ReadFile(source)

	if err != nil {
		return err
	}

	sourceHash := CreateHash(sourcebytes)

	sourcebytes = append(sourcebytes, sourceHash...)

	if err != nil {
		return err
	}

	aesCipher, err := aes.NewCipher(key)
	blowfishCipher, err := blowfish.NewCipher(key)

	if err != nil {
		return err
	}

	iv := make([]byte, 16)
	_, err = rand.Read(iv)

	aesEncrypter := cipher.NewCFBEncrypter(aesCipher, iv)
	blowfishEncrypter := cipher.NewCFBEncrypter(blowfishCipher, iv[0:8])

	destbytes := make([]byte, len(sourcebytes))

	aesEncrypter.XORKeyStream(destbytes, sourcebytes)
	sourcebytes = destbytes
	blowfishEncrypter.XORKeyStream(destbytes, sourcebytes)

	if err != nil {
		return err
	}

	destbytes = append(destbytes, iv...)

	//TODO: Use the same credentials as the source
	ioutil.WriteFile(dst, destbytes, 777)

	return nil
}

func CreateHash(in []byte) []byte {
	h := sha3.NewKeccak256()
	h.Write(in)
	return h.Sum(nil)
}

func DecryptFile(source, dst string, key []byte) error {
	sourcebytes, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	iv := sourcebytes[len(sourcebytes)-16:]
	sourcebytes = sourcebytes[:len(sourcebytes)-16]

	contentLength := len(sourcebytes)
	destbytes := make([]byte, contentLength)

	aesCipher, err := aes.NewCipher(key)

	if err != nil {
		return err
	}

	blowfishCipher, err := blowfish.NewCipher(key)

	blowFishDecryptor := cipher.NewCFBDecrypter(blowfishCipher, iv[0:8])
	blowFishDecryptor.XORKeyStream(destbytes, sourcebytes)
	sourcebytes = destbytes

	aesDecryptor := cipher.NewCFBDecrypter(aesCipher, iv)
	aesDecryptor.XORKeyStream(destbytes, sourcebytes)

	h := destbytes[len(destbytes)-32:]
	destbytes = destbytes[:len(destbytes)-32]

	if !AreEqual(h, CreateHash(destbytes)) {
		return fmt.Errorf("Wrong password")
	}

	//TODO: Use the same credentials as the source
	ioutil.WriteFile(dst, destbytes, 777)

	return nil
}

func AreEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func CreateKeyFromPassword(pwd string) ([]byte, error) {
	if len(pwd) < 8 {
		return nil, fmt.Errorf("Password has to be at least 8 characters")
	}

	h := sha3.NewKeccak256()
	h.Write([]byte(pwd))

	return h.Sum(nil), nil
}

func main() {
	src := "test.txt"
	encrypt := "test.txt.crypt"
	decrypt := "test.txt.decrypt"
	passwd := "TestPasswort"
	key, err := CreateKeyFromPassword(passwd)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = EncryptFile(src, encrypt, key)
	if err != nil {
		fmt.Println(err)
	}

	err = DecryptFile(encrypt, decrypt, key)

	if err != nil {
		fmt.Println(err)
	}
}
