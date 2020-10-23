# FIXME: wait to prevent race condition that makes zypper install to fail
# retriving metadata from repositories
  - while [ $(ps aux | grep zypper | grep -v grep | wc -l) != 0 ]; do sleep 5; done;
  - echo "solver.onlyRequires = true" >> /etc/zypp/zypp.conf
  - zypper --non-interactive --ignore-unknown install ${packages}
