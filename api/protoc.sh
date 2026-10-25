#!/bin/bash

# Compiles the proto files using buf.
set -e

cd api/
buf dep update
buf generate

echo "Proto files compiled successfully using buf."
