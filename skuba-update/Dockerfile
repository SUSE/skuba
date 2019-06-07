FROM registry.opensuse.org/opensuse/leap:latest
MAINTAINER SUSE Containers Team <containers@suse.com>

COPY *requirements.txt /
COPY test/unit-tests-init.sh /init

RUN zypper -n ref && \
    zypper -n in python3-setuptools python3-pip ShellCheck lsof && \
    pip install -r /dev-requirements.txt && \
    chmod +x /init && \
    useradd -m -d /home/suse -s /bin/bash suse && \
    zypper clean -a

USER suse
ENTRYPOINT ["/init"]
