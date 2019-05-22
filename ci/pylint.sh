#!/bin/bash

echo "Starting $0 script"

SDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
VENVDIR=${WORKSPACE:-~}/py3venv
export VENVDIR=${VENVDIR}

setup_python_env() {
    type python3 >/dev/null 2>&1 || sudo zypper install --no-confirm python3

    if [ ! -d $VENVDIR ]; then
        echo "Creating Python 3 virtualenv"
        python3 -m venv $VENVDIR
    fi
    source ${VENVDIR}/bin/activate
    REQ_PATH="${SDIR}/utils/requirements.txt"
    pip3 install -r "${REQ_PATH}" &>/dev/null
    deactivate
}

lint() {
    source ${VENVDIR}/bin/activate
    find . -type f -name "*.py" | grep -v bare-metal | xargs pylint --rcfile="${SDIR}/.pylintrc"
    ret=$?
    deactivate
    exit $ret
}

setup_python_env
lint
