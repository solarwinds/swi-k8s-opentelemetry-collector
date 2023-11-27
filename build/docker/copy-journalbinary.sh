#!/bin/bash

# The binary whose dependencies we want to find
BINARY="/bin/journalctl"

# The directory where we'll copy the dependencies
DEP_DIR="/journalctl-deps"

# Function to copy the binary and its dependencies
copy_dependencies() {
    # Find the dependencies using ldd

    local deps=$(ldd $BINARY | grep "=>" | awk '{print $3}')
    
    # Create the dependency directory if it doesn't exist
    mkdir -p $DEP_DIR

    # Copy the binary itself
    mkdir -p $DEP_DIR/bin
    cp $BINARY $DEP_DIR/bin

    # Copy each dependency
    for dep in $deps; do
        # Create subdirectories if necessary
        local dir=$(dirname $dep)
        mkdir -p $DEP_DIR$dir

        # Copy the library file
        echo "Copying $dep to $DEP_DIR$dir"
        cp $dep $DEP_DIR$dir
    done

    # Copy the dynamic linker
    local linker=$(ldd $BINARY | grep 'ld-linux' | awk '{print $1}')
    if [ -n "$linker" ]; then
        mkdir -p $DEP_DIR$(dirname $linker)

        echo "Copying $linker to $DEP_DIR$(dirname $linker)"
        cp $linker $DEP_DIR$(dirname $linker)
    fi
}

copy_dependencies

echo "Dependencies copied to $DEP_DIR"
