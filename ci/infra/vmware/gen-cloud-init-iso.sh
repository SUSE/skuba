#!/bin/sh

if [[ $# -ne 2 ]] ; then
    echo "Please provide 2 arguments: $0 <worker|master|lb> <cc_yaml_data>"
    exit 1
fi

TMPDIR=$(mktemp -d -p .)

# Look for *.ccfile files in cloud-init dir and use their basename without the extension as files on ISO
for file in cloud-init/*.ccfile; do
    cat $file > $TMPDIR/$(basename -- "$file" .ccfile)
done

# TODO process yaml data in ${@:2} variable to not contain any non-pair single or double quotes
echo -e "${@:2}" > $TMPDIR/user-data

mkisofs -output cc-"$1".iso -volid cidata -joliet -rock $TMPDIR

rm -r $TMPDIR
