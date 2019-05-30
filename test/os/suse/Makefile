MAKEFILE_PATH := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

OPENSUSE_LEAP_IMAGE = registry.opensuse.org/opensuse/leap
OPENSUSE_TUMBLEWEED_IMAGE = registry.opensuse.org/opensuse/tumbleweed
SLE_15_IMAGE = registry.suse.de/suse/sle-15-sp1/update/cr/images/suse/sle15:15.1
BUILDER_IMAGE = opensuse/tumbleweed:build-rpm

.PHONY: test
test: repos test-opensuse test-sle

.PHONY: test-opensuse
test-opensuse: test-opensuse-tumbleweed test-opensuse-leap

.PHONY: test-opensuse-tumbleweed
test-opensuse-tumbleweed:
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_TUMBLEWEED_IMAGE) /suse/test-interruptive-updates.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_TUMBLEWEED_IMAGE) /suse/test-non-interruptive-updates.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_TUMBLEWEED_IMAGE) /suse/test-interruptive-updates-with-needreboot.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_TUMBLEWEED_IMAGE) /suse/test-non-interruptive-updates-with-needreboot.sh

.PHONY: test-opensuse-leap
test-opensuse-leap:
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_LEAP_IMAGE) /suse/test-interruptive-updates.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_LEAP_IMAGE) /suse/test-non-interruptive-updates.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_LEAP_IMAGE) /suse/test-interruptive-updates-with-needreboot.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(OPENSUSE_LEAP_IMAGE) /suse/test-non-interruptive-updates-with-needreboot.sh

.PHONY: test-sle
test-sle: test-sle-15

.PHONY: test-sle-15
test-sle-15:
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(SLE_15_IMAGE) /suse/test-interruptive-updates.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(SLE_15_IMAGE) /suse/test-non-interruptive-updates.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(SLE_15_IMAGE) /suse/test-interruptive-updates-with-needreboot.sh
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(SLE_15_IMAGE) /suse/test-non-interruptive-updates-with-needreboot.sh

.PHONY: rpm-builder-image
rpm-builder-image:
	docker build -t $(BUILDER_IMAGE) -f Dockerfile.builder .

.PHONY: rpms
rpms: artifacts/caasp-test-1-1.noarch.rpm artifacts/caasp-test-2-1.noarch.rpm

artifacts/%.rpm: rpm-builder-image
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(BUILDER_IMAGE) /suse/build-rpm.sh $*

.PHONY: repos
repos: rpms artifacts/repo-base artifacts/repo-update-with-reboot-suggested artifacts/repo-update-without-reboot-suggested

artifacts/repo-%:
	docker run --rm -v $(MAKEFILE_PATH):/suse -it $(BUILDER_IMAGE) /suse/build-repo.sh $*

.PHONY: clean
clean:
	rm -rf artifacts/base-repo
	rm -rf artifacts/update-repo-with-reboot-suggested
	rm -rf artifacts/update-repo-without-reboot-suggested
	rm -f artifacts/*.rpm
