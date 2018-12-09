# fcompare - A tiny file comparison program written in go

This is a quick way for locating identical files in a folder.
Pass to the program the folder you want to check

Build

```bash
go get -u github.com/omerh/fcompare
# If GOPATH is set
cd $GOPATH/src/github.com/omerh/fcompare
# If no GOPATH is set
cd ~/go/src/github.com/omerh/fcompare
go build
```

Run the app

```bash
./fcompare /directory
```
