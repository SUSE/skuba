#!/bin/sh
#
# LXD image build
#

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

FIRST_GOPATH=$(echo `echo ${GOPATH} | cut -f1 -d':'`)
GOBIN=$([ -n "${GOBIN}" ] && echo ${GOBIN} || (echo $(GOPATH)/bin))

DISTROBUILDER=
DISTROBUILDER_YAML=$DIR/distrobuilder-opensuse.yaml

IMAGE_ALIAS="opensuse-caasp"

FORCE=

read -r -d '' HELP <<- EOM
Usage:

  $0 [args]

where [args] can be:

  --img <IMAGE>          the name for the image to load in LXD
                         (default: $IMAGE_ALIAS)
  --yaml <YAML>          the YAML definition file
                         (default: $DISTROBUILDER_YAML)
  --distrobuilder <EXE>  the distrobuilder executable
                         (will be downloaded if no provided/autodetected)
  --force <BOOL>         force the image (re)creation

EOM

while [ $# -gt 0 ] ; do
  case $1 in
    --img|--alias)
      IMAGE_ALIAS=$2
      shift
      ;;
    --distrobuilder|--exe)
      DISTROBUILDER=$2
      shift
      ;;
    --yaml|--definition)
      DISTROBUILDER_YAML=$2
      shift
      ;;
    --force)
      case $2 in
        true|TRUE|yes|YES|1)
          FORCE=1
          ;;
        false|FALSE|no|NO|0)
          FORCE=
          ;;
      esac
      shift
      ;;
    --help|-h)
      echo "$HELP"
      exit 0
      ;;
    *)
      echo "unknown argument $1"
      exit 1
      ;;
  esac
  shift
done

#########################################################################

check_distrobuilder() {
    if [ -n "$DISTROBUILDER" ] ; then
        echo ">>> (distrobuilder provided: $DISTROBUILDER)"
        return
    fi

    if command distrobuilder &>/dev/null ; then
        DISTROBUILDER="$(which distrobuilder)"
        echo ">>> (distrobuilder found in the PATH)"
        return
    fi

    if [ -x $GOBIN/distrobuilder ] ; then
        echo ">>> (distrobuilder found at $GOBIN/distrobuilder)"
        DISTROBUILDER=$GOBIN/distrobuilder
        return
    fi

    echo ">>> downloading distrobuilde with 'go get'..."
    go get -d github.com/lxc/distrobuilder
    echo ">>> building distrobuilder..."
    make -C $FIRST_GOPATH/src/github.com/lxc/distrobuilder
    echo ">>> distrobuilder installed at $GOBIN/distrobuilder"
    DISTROBUILDER=$GOBIN/distrobuilder
}

#########################################################################
# LXD images

image_exists() {
    lxc image show $IMAGE_ALIAS &>/dev/null
}

image_delete() {
    echo ">>> deleting any previous image $IMAGE_ALIAS..."
    lxc image delete $IMAGE_ALIAS 2>/dev/null || /bin/true
}

image_artifacts_exist() {
    [ -f lxd.tar.xz ] && [ -f rootfs.squashfs ]
}

image_build() {
    echo ">>> building LXD image with $DISTROBUILDER..."
    echo ">>> IMPORTANT: distrobuilder will be launched with SUDO !!!"
    sudo $DISTROBUILDER --cleanup=false build-lxd $DISTROBUILDER_YAML
}

image_import() {
    image_delete

    echo ">>> importing LXD image..."
    lxc image import lxd.tar.xz rootfs.squashfs --alias $IMAGE_ALIAS
    echo ">>> removing leftovers"
    rm -f lxd.tar.xz rootfs.squashfs
    echo ">>> current list of LXD images:"
    lxc image list
}

image_cleanup() {
    echo ">>> cleaning up leftovers..."
    # just in case...
}

[ -n "$FORCE" ] && image_delete
echo ">>> checking if $IMAGE_ALIAS image exists..."
if ! image_exists ; then
    echo ">>> $IMAGE_ALIAS does not exist: building..."
    check_distrobuilder
    trap image_cleanup EXIT
    image_artifacts_exist || image_build
    image_artifacts_exist && image_import
else
    echo ">>> $IMAGE_ALIAS already present!"
fi

