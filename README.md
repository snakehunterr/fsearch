```
    dMMMMMP .dMMMb  dMMMMMP .aMMMb  dMMMMb  .aMMMb  dMP dMP
   dMP     dMP" VP dMP     dMP"dMP dMP.dMP dMP"VMP dMP dMP
  dMMMP    VMMMb  dMMMP   dMMMMMP dMMMMK" dMP     dMMMMMP
 dMP     dP .dMP dMP     dMP dMP dMP"AMF dMP.aMP dMP dMP
dMP      VMMMP" dMMMMMP dMP dMP dMP dMP  VMMMP" dMP dMP
```

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/snakehunterr/fsearch)
![GitHub License](https://img.shields.io/github/license/snakehunterr/fsearch)
![GitHub Version](<https://img.shields.io/github/v/tag/snakehunterr/fsearch?include_prereleases&sort=date&label=version&color=hex(2343ca)>)

A tool to find files, written in golang with zero dependency and enhanced with parallel power.

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
