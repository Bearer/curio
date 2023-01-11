#!/bin/sh -

DEBIAN_RELEASES=`debian-distro-info --supported`
UBUNTU_RELEASES=`(ubuntu-distro-info --supported-esm; ubuntu-distro-info --supported) | sort -u`

cd curio-repo/deb

for release in $DEBIAN_RELEASES $UBUNTU_RELEASES
do
  echo "Removing deb package of $release"
  for arch in i386 amd64 arm64
  do
    reprepro -A $arch remove $release curio
  done
done

for release in $DEBIAN_RELEASES $UBUNTU_RELEASES
do
  echo "Adding deb package to $release"
  for arch in 64bit 32bit ARM64
  do
    reprepro includedeb $release ../../dist/*Linux-$arch.deb
  done
done

git add .
git commit -m "Update deb packages"
git push origin main
