FROM registry.opensuse.org/opensuse/tumbleweed

RUN zypper ref && zypper -n in rpm-build rpmdevtools createrepo libcreaterepo_c-devel
RUN rm /var/run/reboot-needed
