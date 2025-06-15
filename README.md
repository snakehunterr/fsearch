# FSEARCH

FSEARCH (fs) is a tool to find files, written in golang with zero dependency and enhanced with parallel power.

## Current supported platforms

- MacOS
- Linux

**_Why?_**  
Because under the hood **fsearch** directly using **syscall** go package to travel directories and
get information about files: depending on OS, the file info varies.

## Installing

```sh
git clone "github.com/snakehunterr/fsearch"
cd fsearch
make build
# then ./fs ....
# you also can:
# sudo mv ./fs /usr/bin/fs
```

## Usage

```sh
fs -help # show help and exit

fs . # simply travel CWD and print out all files
fs # same as above (by default, PATH is '.')

# -name is a regex filter, executed on each file name;
fs -name '.*.go' ~
# or
fs ~ -name '.*.go'

# print files, which names ending with '.go' and do not contain 'main'
fs -name '\.go$' -iname 'main' ~
```
