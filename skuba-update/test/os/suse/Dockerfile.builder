FROM registry.opensuse.org/opensuse/leap:15.2

RUN zypper ref && zypper -n in rpm-build rpmdevtools createrepo libcreaterepo_c-devel
RUN rm /var/run/reboot-needed
