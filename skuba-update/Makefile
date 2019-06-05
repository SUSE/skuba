.PHONY: all
all: install

.PHONY: install
install:
	./setup.py install

.PHONY: suse-package
suse-package:
	ci/packaging/suse/rpmfiles_maker.sh

.PHONY: suse-changelog
suse-changelog:
	ci/packaging/suse/changelog_maker.sh "$(CHANGES)"

.PHONY: test
test:
	tox
