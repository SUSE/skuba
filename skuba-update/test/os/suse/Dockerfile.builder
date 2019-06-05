FROM registry.opensuse.org/opensuse/tumbleweed

RUN zypper ref && zypper -n in rpm-build rpmdevtools createrepo
RUN rm /var/run/reboot-needed