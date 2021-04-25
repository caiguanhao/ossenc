package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
)

type (
	config struct {
		EncryptionKey      key
		FileNameFormat     string
		OSSAccessKeyId     string
		OSSAccessKeySecret string
		OSSPrefix          string
		OSSBucket          string
	}

	key []byte
)

func (k key) String() string {
	return hex.EncodeToString(k)
}

func (k key) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(k))
}

func (k *key) UnmarshalJSON(data []byte) (err error) {
	var src string
	err = json.Unmarshal(data, &src)
	if err != nil {
		return
	}
	*k, err = hex.DecodeString(src)
	return err
}

func updateConfigFile(file string, c *config) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln(err)
	}
	e := json.NewEncoder(f)
	e.SetIndent("", "\t")
	k := make(key, 32)
	rand.Read(k)
	var toCreate bool
	if c == nil {
		c = &config{
			EncryptionKey: k,
		}
		toCreate = true
	}
	err = e.Encode(c)
	if err != nil {
		log.Fatalln(err)
	}
	if toCreate {
		log.Println("Config file created:", file)
	} else {
		log.Println("Config file updated:", file)
	}
	return
}
