# How to install libvips on Windows

## Windows

+ Install [msys2-x86_64](http://www.msys2.org/
+ In msys shell, do:
```
pacman -Syuu
pacman -S mingw-w64-x86_64-gcc
pacman -S mingw-w64-x86_64-pkg-config
pacman -S mingw-w64-x86_64-zlib
pacman -S mingw-w64-x86_64-libvips
pacman -S mingw-w64-x86_64-poppler
pacman -S mingw-w64-x86_64-libjxl
pacman -S mingw-w64-x86_64-openslide
```

+ Switch to cmd.exe shell, and do:
```
set PATH=<msys64>\mingw64\bin;%PATH%
set PKG_CONFIG_PATH=<msys64>\mingw64\lib\pkgconfig
```

The default installation path of `msys2` is `C:\msys64` and you must change
`<msys64>` to your installation path of `msys2`.

## Windows Alternative
In case you are having struggles running libvips on Windows through the recommended instructions, [**Golibvips**](https://github.com/twincats/golibvips/ is alternative libvips wrapper that work with windows by using vips Windows binary

### Installation

#### Step 1 : Download Vips windows binary

Download vips windows binary from here [vips-dev-w64-all-8.13.3.zip](https://github.com/libvips/build-win64-mxe/releases/download/v8.13.3/vips-dev-w64-all-8.13.3.zip

#### Step 2 : Setting Environtment Variable

extract `vips-dev-w64-all-8.13.3.zip` to folder, for example C:\vips-dev-8.13
windows search Edit the system environtment variable > Environtment Variables > Path > edit > new > type `C:\vips-dev-8.13\bin`.

#### Step 3 : Install Golibvips golang package

``` 
go get -u github.com/twincats/golibvips/libvips 
```

## Common

Check if pkg-config is able to find the right ImageMagick include and libs:

```
pkg-config --cflags --libs libvips
```

Then go get it:

```
go get -u github.com/davidbyttow/govips/v2/vips
```

Per the security update https://groups.google.com/forum/#!topic/golang-announce/X7N1mvntnoU
you may need whitelist the -Xpreprocessor flag in your environment.

```
export CGO_CFLAGS_ALLOW='-Xpreprocessor'
```