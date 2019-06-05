PWD := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

OPENSUSE_LEAP_IMAGE = registry.opensuse.org/opensuse/leap:latest
OPENSUSE_TUMBLEWEED_IMAGE = registry.opensuse.org/opensuse/tumbleweed:latest
SLE_15_IMAGE = registry.suse.de/suse/sle-15-sp1/update/cr/images/suse/sle15:15.1
BUILDER_IMAGE = opensuse/tumbleweed:build-rpm

.PHONY: test
test: test-opensuse test-sle

.PHONY: test-opensuse
test-opensuse: test-opensuse-tumbleweed test-opensuse-leap

.PHONY: test-opensuse-tumbleweed
test-opensuse-tumbleweed: repos
	$(PWD)/run_tests.sh $(OPENSUSE_TUMBLEWEED_IMAGE)

.PHONY: test-opensuse-leap
test-opensuse-leap: repos
	$(PWD)/run_tests.sh $(OPENSUSE_LEAP_IMAGE)

.PHONY: test-sle
test-sle: test-sle-15

.PHONY: test-sle-15
test-sle-15: repos
	$(PWD)/run_tests.sh $(SLE_15_IMAGE)

.PHONY: rpm-builder-image
rpm-builder-image:
	pushd $(PWD) >/dev/null
	docker build -t $(BUILDER_IMAGE) -f $(PWD)/Dockerfile.builder $(PWD)

.PHONY: rpms
rpms: artifacts/caasp-test-1-1.noarch.rpm artifacts/caasp-test-2-1.noarch.rpm

artifacts/%.rpm: rpm-builder-image
	docker run --rm -v $(PWD):/suse -it $(BUILDER_IMAGE) /suse/build-rpm.sh $*

.PHONY: repos
repos: rpms artifacts/repo-base artifacts/repo-update-with-reboot-suggested artifacts/repo-update-with-restart-suggested artifacts/repo-update-plain artifacts/repo-update-with-message

artifacts/repo-%:
	docker run --rm -v $(PWD):/suse -it $(BUILDER_IMAGE) /suse/build-repo.sh $*

.PHONY: clean
clean:
	rm -rf artifacts/base-repo
	rm -rf artifacts/update-repo-with-reboot-suggested
	rm -rf artifacts/update-repo-without-reboot-suggested
	rm -f artifacts/*.rpm
