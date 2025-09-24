// encryptor.go
package main

import (
	"crypto/rc4"
	"io/ioutil"
	"os"
	"fmt"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: encryptor <raw_shellcode.bin> <key.txt> <out_encrypted.bin>")
		return
	}
	data, _ := ioutil.ReadFile(os.Args[1])
	key, _ := ioutil.ReadFile(os.Args[2])
	cipher, _ := rc4.NewCipher(key)
	dst := make([]byte, len(data))
	cipher.XORKeyStream(dst, data)
	ioutil.WriteFile(os.Args[3], dst, 0644)
	fmt.Printf("Encrypted %d bytes with key of length %d\n", len(data), len(key))
}

