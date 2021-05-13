package main

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"

	"github.com/gopsql/goconf"
)

type (
	config struct {
		EncryptionKey key `
Use hexadecimal string with 64 characters to express 256-bit encryption key for
aes-256-ofb. To generate new key, leave this key empty, save this file and then
run ossenc -C.`
		FileNameFormat string `
Format of the remote file name for upload, %{name} for the file name without
extension, %{ext} for the file extension. You can also use formats used in
strftime(), for example, %Y (year), %m (month), %d (day), %H (hour), %M
(minute), %S (second), %s (timestamp), %N (nanosecond).`
		OSSAccessKeyId string `
Access key ID for Aliyun OSS.`
		OSSAccessKeySecret string `
Access key secret for Aliyun OSS.`
		OSSPrefix string `
URL prefix for Aliyun OSS. https://<bucket-name>.<region>.aliyuncs.com/<root>`
		OSSBucket string `
Bucket name for Aliyun OSS.`
	}

	key []byte
)

func (k *key) SetString(input string) (err error) {
	*k, err = hex.DecodeString(input)
	return
}

func (k key) String() string {
	return hex.EncodeToString(k)
}

func readConf(file string, toUpdate bool) (conf config) {
	created := false
	content, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) && toUpdate {
			conf = config{
				FileNameFormat: "%Y/%m/%d/%{name}%{ext}",
			}
			created = true
		} else {
			log.Fatalln(err)
		}
	} else {
		err = goconf.Unmarshal([]byte(content), &conf)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if toUpdate {
		if len(conf.EncryptionKey) == 0 {
			k := make(key, 32)
			rand.Read(k)
			conf.EncryptionKey = k
		}
		content, err := goconf.Marshal(conf)
		if err != nil {
			log.Fatalln(err)
		}
		err = ioutil.WriteFile(file, content, 0600)
		if err != nil {
			log.Fatalln(err)
		}
		if created {
			log.Println("Config file created:", file)
		} else {
			log.Println("Config file updated:", file)
		}
		os.Exit(0)
	}
	return
}
