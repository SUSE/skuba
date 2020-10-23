# FIXME: wait to prevent race condition that makes zypper install to fail
# retriving metadata from repositories
  - sleep 30
  - zypper --non-interactive --ignore-unknown install ${packages}
