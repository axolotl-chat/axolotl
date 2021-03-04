#!/bin/bash

# copy click files
echo "copy click files $@"

cp -a ../../../click/* $@
# Build axolotl-web
echo "update translations and build axolotl-web $@"
cd ../../../axolotl-web&&npm run translate &&npm run build && mkdir $@/axolotl-web&&cp dist $@/axolotl-web/ -r
