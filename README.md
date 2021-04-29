# ossenc

Upload, download, list and delete encrypted files on Aliyun OSS. Files are
compressed using zlib, encrypted using aes-256-ofb.

```
go get -v -u github.com/caiguanhao/ossenc
```

By default, remote file deletion function (the -D option) is not included, to
enable it, you must build with `delete` tag:

```
go get -v -u -tags delete github.com/caiguanhao/ossenc
```

Options:

```
Usage of ossenc:
  -C	create (update if exists) config file and exit
  -D	delete remote files
  -F	do not format remote file name, ignore FileNameFormat config
  -O	just like -o but use remote file name
  -P	print openssl decryption command after upload
  -c string
    	location of the config file (default "~/.ossenc.go")
  -f string
    	file name format, overrides FileNameFormat config
  -l	list directory
  -n	dry-run, do not upload or download any file
  -o string
    	download remote file to local file, use - for stdout
  -p	do not show progress
```

## Config

You must enter Aliyun OSS API key ID and secret string in the config file,
default location of the config file is `~/.ossenc.go`.

```
# create config
ossenc -C
```

### EncryptionKey

The 256-bit encryption key for aes-256-ofb.

### FileNameFormat

Format of the remote file name for upload and download.

`%{name}` - file name without extension

`%{ext}` - file extension

You can also use formats used in
[strftime()](https://github.com/caiguanhao/strftime), for example, `%Y` (year),
`%m` (month), `%d` (day), `%H` (hour), `%M` (minute), `%S` (second), `%s`
(timestamp), `%N` (nanosecond).

### OSSAccessKeyId

Access key ID for Aliyun OSS.

### OSSAccessKeySecret

Access key secret for Aliyun OSS.

### OSSPrefix

URL prefix for Aliyun OSS. `https://<bucket-name>.<region>.aliyuncs.com/<root>`.

### OSSBucket

Bucket name for Aliyun OSS.

## Upload

Default action of `ossenc`.

```
# upload multiple local files to remote
ossenc localFiles...

# pipe
cat localFile | ossenc
```

### Download

```
# download to localFile
ossenc -o localFile remoteFile

# download to local, use same name
ossenc -O remoteFile

# output to stdin
ossenc -o - remoteFile
```

### List

```
# list contents of current directory
ossenc -l

# list contents of root
ossenc -l -F /
```

### Delete

You must build `ossenc` with tag `delete`.

```
ossenc -D remoteFiles...
```
