package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type (
	config struct {
		EncryptionKey key `
Use hexadecimal string with 64 characters to express 256-bit encryption key for
aes-256-ofb.`
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

func unmarshalConfig(input []byte, v interface{}) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", string(input), 0)
	if err != nil {
		return err
	}
	rv := reflect.ValueOf(v)
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok != token.CONST {
			continue
		}
		for _, sp := range gd.Specs {
			vs, ok := sp.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, v := range vs.Values {
				bl, ok := v.(*ast.BasicLit)
				if !ok {
					continue
				}
				value, err := strconv.Unquote(bl.Value)
				if err != nil {
					return err
				}
				for _, name := range vs.Names {
					field := rv.Elem().FieldByName(name.Name)
					if field.Kind() == reflect.String {
						field.SetString(value)
					} else if i, ok := field.Addr().Interface().(interface{ SetString(string) error }); ok {
						err := i.SetString(value)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func marshalConfig(c config) (output string) {
	output = `package config

const (
`
	rt := reflect.TypeOf(c)
	rv := reflect.ValueOf(c)
	for i := 0; i < rt.NumField(); i++ {
		if i > 0 {
			output += "\n"
		}
		if tag := string(rt.Field(i).Tag); tag != "" {
			lines := strings.Split(strings.TrimSpace(tag), "\n")
			for _, line := range lines {
				output += "\t// " + line + "\n"
			}
		}
		v := fmt.Sprint(rv.Field(i).Interface())
		output += "\t" + rt.Field(i).Name + " = " + strconv.Quote(v) + "\n"
	}
	output += `)
`
	return
}

func readConf(file string, toUpdate bool) (conf config) {
	created := false
	content, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) && toUpdate {
			k := make(key, 32)
			rand.Read(k)
			conf = config{
				EncryptionKey:  k,
				FileNameFormat: "%Y/%m/%d/%{name}%{ext}",
			}
			created = true
		} else {
			log.Fatalln(err)
		}
	} else {
		err = unmarshalConfig([]byte(content), &conf)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if toUpdate {
		err = ioutil.WriteFile(file, []byte(marshalConfig(conf)), 0600)
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
