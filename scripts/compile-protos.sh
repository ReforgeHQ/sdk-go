#!/bin/bash

#run from the root directory

# Copy prefab.proto to reforge.proto to avoid proto registry conflicts
# with prefab-cloud-go when both are used in the same binary
cp proto-source/prefab.proto proto-source/reforge.proto

protoc --proto_path=proto-source \
  --go_out=proto --go_opt=paths=source_relative \
  --go_opt=Mproto-source/reforge.proto=github.com/ReforgeHQ/sdk-go/proto \
  proto-source/reforge.proto

# Clean up the temporary copy
rm proto-source/reforge.proto
