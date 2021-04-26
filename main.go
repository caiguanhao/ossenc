package main

import (
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/caiguanhao/ossslim"
	"github.com/caiguanhao/strftime"
	"golang.org/x/term"
)

const (
	command = "curl -s %s | openssl enc -d -aes-256-ofb -iv 0 -K %s | unpigz\n"
)

var (
	noProgress   bool
	dryRun       bool
	printCommand bool

	conf   config
	client ossslim.Client
)

func main() {
	defaultConfigFile := ".ossenc.go"
	if home, _ := os.UserHomeDir(); home != "" {
		defaultConfigFile = filepath.Join(home, defaultConfigFile)
	}

	flag.BoolVar(&noProgress, "p", false, "do not show progress")
	flag.BoolVar(&printCommand, "P", false, "print decryption command")
	configFile := flag.String("c", defaultConfigFile, "location of the config file")
	createConfig := flag.Bool("C", false, "create (update if exists) config file and exit")
	format := flag.String("F", "", "file name format, overrides FileNameFormat config")
	listDirectory := flag.Bool("l", false, "list directory")
	output := flag.String("o", "", "download remote file to local file, use - for stdout")
	outputRemote := flag.Bool("O", false, "just like -o but use remote file name")
	flag.BoolVar(&dryRun, "n", false, "dry-run, do not upload")
	flag.Parse()

	conf = readConf(*configFile, *createConfig)

	if *format != "" {
		conf.FileNameFormat = *format
	}

	p, err := url.Parse(conf.OSSPrefix)
	if err != nil {
		log.Fatalln(err)
		return
	}
	dir := p.Path
	p.Path = ""
	p.RawQuery = ""
	p.Fragment = ""

	client = ossslim.Client{
		AccessKeyId:     conf.OSSAccessKeyId,
		AccessKeySecret: conf.OSSAccessKeySecret,
		Prefix:          p.String(),
		Bucket:          conf.OSSBucket,
	}

	if *listDirectory {
		var prefix string
		if len(flag.Args()) > 0 {
			prefix = flag.Arg(0)
		} else {
			prefix = dir
		}
		list(prefix)
		return
	}

	if *outputRemote {
		if len(flag.Args()) == 0 {
			log.Fatalln("You must provide remote file name.")
		}
		b := filepath.Base(flag.Arg(0))
		output = &b
	}
	if *output != "" {
		if len(flag.Args()) == 0 {
			log.Fatalln("You must provide remote file name.")
		}
		var target io.Writer
		if *output == "-" {
			target = os.Stdout
			noProgress = true
		} else {
			f, err := os.OpenFile(*output, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				log.Fatalln(err)
			}
			defer f.Close()
			target = f
		}
		download(target, flag.Arg(0), *output)
		return
	}

	if len(flag.Args()) == 0 {
		dest := path.Join(dir, formatName("stdin"))
		if term.IsTerminal(int(os.Stdin.Fd())) {
			noProgress = true
			fmt.Fprintln(os.Stderr, "Input content and press Ctrl-D to finish or Ctrl-C to abort:")
		}
		upload(os.Stdin, nil, "stdin", dest)
		return
	}

	for _, arg := range flag.Args() {
		f, err := os.Open(arg)
		if err != nil {
			log.Fatalln(err)
		}
		var total *int64
		if fi, err := f.Stat(); err == nil {
			s := fi.Size()
			total = &s
		}
		dest := path.Join(dir, formatName(arg))
		upload(f, total, arg, dest)
		f.Close()
	}

}

func list(prefix string) {
	result, err := client.List(prefix, true)
	if err != nil {
		log.Fatalln(err)
	}
	nameLen := 20
	sizeLen := 1
	for _, f := range result.Files {
		nl := len(f.Name)
		if nl > nameLen {
			nameLen = nl
		}
		sl := len(strconv.FormatInt(f.Size, 10))
		if sl > sizeLen {
			sizeLen = sl
		}
	}
	for _, f := range result.Files {
		fmt.Printf(fmt.Sprintf("%%-%ds\t%%%dd\t%%s\n", nameLen, sizeLen), f.Name, f.Size, f.LastModified)
	}
}

func upload(reader io.Reader, total *int64, src, path string) {
	if dryRun {
		if printCommand {
			fmt.Println("#", src)
			fmt.Printf(command, client.URL(path), conf.EncryptionKey)
		} else {
			fmt.Println(src, "->", client.URL(path))
		}
		return
	}

	r, w := io.Pipe()

	defer r.Close()

	go func() {
		defer w.Close()

		prog := &progress{
			name:  filepath.Base(path),
			total: total,
		}

		var writer io.Writer
		if noProgress == false {
			go prog.Run()
			writer = io.MultiWriter(w, prog)
		} else {
			writer = w
		}

		if err := compress(conf.EncryptionKey, reader, writer); err != nil {
			log.Fatalln(err)
		}

		if noProgress == false {
			prog.Close()
			time.Sleep(1500 * time.Millisecond) // time for the final message
		}
	}()

	req, err := client.Upload(path, r, nil, "")
	if err != nil {
		log.Fatalln(err)
		return
	}

	if printCommand {
		fmt.Printf(command, req.URL(), conf.EncryptionKey)
	} else {
		fmt.Fprintln(os.Stderr, src, "->", client.URL(path))
	}
}

func download(target io.Writer, key, dest string) {
	r, w := io.Pipe()

	defer r.Close()
	defer w.Close()

	prog := &progress{
		name: filepath.Base(key),
	}

	var writer io.Writer
	if noProgress == false {
		go prog.Run()
		writer = io.MultiWriter(w, prog)
	} else {
		writer = w
	}
	req, err := client.DownloadAsync(key, writer)
	if err != nil {
		log.Fatalln(err)
	}
	prog.total = req.ResponseContentLength

	err = decompress(conf.EncryptionKey, r, target)
	if err != nil {
		log.Fatalln(err)
	}

	if noProgress == false {
		prog.Close()
		time.Sleep(1500 * time.Millisecond) // time for the final message
	}

	if noProgress == false {
		fmt.Fprintln(os.Stderr, req.URL(), "->", dest)
	}
}

func compress(key key, reader io.Reader, writer io.Writer) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	var iv [aes.BlockSize]byte
	cipherW := &cipher.StreamWriter{
		S: cipher.NewOFB(block, iv[:]),
		W: writer,
	}
	gzipW := zlib.NewWriter(cipherW)
	if _, err := io.Copy(gzipW, reader); err != nil {
		return err
	}
	if err := gzipW.Close(); err != nil {
		return err
	}
	if err := cipherW.Close(); err != nil {
		return err
	}
	return nil
}

func decompress(key key, reader io.Reader, writer io.Writer) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	var iv [aes.BlockSize]byte
	cipherR := &cipher.StreamReader{
		S: cipher.NewOFB(block, iv[:]),
		R: reader,
	}
	gzipR, err := zlib.NewReader(cipherR)
	if err != nil {
		return err
	}
	if _, err := io.Copy(writer, gzipR); err != nil {
		return err
	}
	if err := gzipR.Close(); err != nil {
		return err
	}
	return nil
}

func formatName(path string) string {
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	base := filename[0 : len(filename)-len(ext)]
	name := strftime.Format(conf.FileNameFormat, time.Now())
	name = strings.ReplaceAll(name, "%{name}", base)
	name = strings.ReplaceAll(name, "%{ext}", ext)
	if name == "" {
		return filename
	}
	return name
}
