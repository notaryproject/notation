#!/bin/sh

# TODO: it would be helpful if this used gosu to drop to the uid/gid of the volume mount before running each make

cd /nv2
make build

cd /oras
make build

cp /nv2/bin/* /usr/local/bin/
cp /oras/bin/linux/amd64/oras /usr/local/bin/
