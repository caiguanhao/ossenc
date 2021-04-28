# ossenc

Upload and download encrypted files to Aliyun OSS.

```
go get -v -u github.com/caiguanhao/ossenc
```

If you want to enable the remote file deletion (-D option):

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
