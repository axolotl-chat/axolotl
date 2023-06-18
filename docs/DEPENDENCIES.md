# Zkgroup

The textsecure-cmd depends on libzkgroup a rust lib.

## Installing

In order to run textsecure-cmd you can either copy the precompiled library with `make copy-lib` or compile it with `make`.

## Troubleshooting

### Cannot find library

```
/tmp/go-build146229246/b001/exe/textsecure: error while loading shared libraries: libzkgroup_linux_x86_64.so: cannot open shared object file: No such file or directory
```

Ensure that your libzkgroup is installed. Either add the current directory to the `LD_LIBRARY_PATH=$(PWD) go run .` or `make run`

To install the library to system run `sudo make install-zkgroup`
