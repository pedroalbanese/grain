package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"log"
	"os"

	"github.com/pedroalbanese/lwcrypto/grain"
)

var (
	aad    = flag.String("a", "", "Additional Associated Data.")
	dec    = flag.Bool("d", false, "Decrypt instead of Encrypt.")
	file   = flag.String("f", "", "Target file. ('-' for STDIN)")
	iter   = flag.Int("i", 1024, "Iterations. (for PBKDF2)")
	key    = flag.String("k", "", "128-bit key to Encrypt/Decrypt.")
	pbkdf  = flag.String("p", "", "PBKDF2.")
	random = flag.Bool("r", false, "Generate random 128-bit cryptographic key.")
	salt   = flag.String("s", "", "Salt. (for PBKDF2)")
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Grain128a Encryption Tool - ALBANESE Research Lab (c) 2020-2022")
		fmt.Fprintln(os.Stderr, "Lightweight AEAD stream cipher submitted to the CAESAR competition\n")
		fmt.Fprintln(os.Stderr, "Usage of "+os.Args[0]+":")
		fmt.Fprintln(os.Stderr, os.Args[0]+" [-d] -p \"pass\" [-i N] [-s \"salt\"] -f <file.ext>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *random == true {
		var key []byte
		var err error
		key = make([]byte, 16)
		_, err = io.ReadFull(rand.Reader, key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(hex.EncodeToString(key))
		os.Exit(0)
	}

	var keyHex string
	var prvRaw []byte
	if *pbkdf != "" {
		prvRaw = pbkdf2.Key([]byte(*pbkdf), []byte(*salt), *iter, 16, sha256.New)
		keyHex = hex.EncodeToString(prvRaw)
	} else {
		keyHex = *key
	}
	var key []byte
	var err error
	if keyHex == "" {
		key = make([]byte, 16)
		_, err = io.ReadFull(rand.Reader, key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(os.Stderr, "Key=", hex.EncodeToString(key))
	} else {
		key, err = hex.DecodeString(keyHex)
		if err != nil {
			log.Fatal(err)
		}
		if len(key) != 16 {
			log.Fatal(err)
		}
	}

	buf := bytes.NewBuffer(nil)
	var data io.Reader
	if *file == "-" {
		data = os.Stdin
	} else {
		data, _ = os.Open(*file)
	}
	io.Copy(buf, data)
	msg := buf.Bytes()

	aead, err := grain.New(key)
	if err != nil {
		panic(err)
	}

	if *dec == false {
		nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead())

		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			log.Fatal(err)
		}

		out := aead.Seal(nonce, nonce, msg, []byte(*aad))
		fmt.Printf("%s", out)

		os.Exit(0)
	}

	if *dec == true {
		nonce, msg := msg[:aead.NonceSize()], msg[aead.NonceSize():]

		out, err := aead.Open(nil, nonce, msg, []byte(*aad))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s", out)

		os.Exit(0)
	}
}
