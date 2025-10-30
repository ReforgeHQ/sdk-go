#!/bin/bash

#run from the root directory

# Copy prefab.proto to reforge.proto to avoid proto registry conflicts
# with prefab-cloud-go when both are used in the same binary.
# We rename both the file and the protobuf package namespace to avoid
# conflicts in the global protobuf type registry.
cp proto-source/prefab.proto proto-source/reforge.proto

# Change the protobuf package from "prefab" to "reforge" to avoid
# namespace collision on types like prefab.OnFailure
sed -i.bak 's/^package prefab;/package reforge;/' proto-source/reforge.proto
rm proto-source/reforge.proto.bak

protoc --proto_path=proto-source \
  --go_out=proto --go_opt=paths=source_relative \
  --go_opt=Mproto-source/reforge.proto=github.com/ReforgeHQ/sdk-go/proto \
  proto-source/reforge.proto

# Clean up the temporary copy
rm proto-source/reforge.proto
