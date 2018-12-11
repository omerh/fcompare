# fcompare - A tiny file comparison program written in go

This is a quick way for locating identical files in a folder.

Pass the folder you want to check to the program

## Build

```bash
go get -u github.com/omerh/fcompare
# If GOPATH is set
cd $GOPATH/src/github.com/omerh/fcompare
# If no GOPATH is set
cd ~/go/src/github.com/omerh/fcompare
go build
```

## Run

```bash
fcompare [-t] <directory>
```

## Args
arg name | type | default value | usage
--- | --- | --- | ---
t | bool | false | set to true for calculating hashes in parallel
