.PHONY: release major minor patch current-version

# Get the latest tag, default to v0.0.0 if no tags exist
LATEST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
MAJOR := $(shell echo $(LATEST_TAG) | cut -d. -f1 | tr -d 'v')
MINOR := $(shell echo $(LATEST_TAG) | cut -d. -f2)
PATCH := $(shell echo $(LATEST_TAG) | cut -d. -f3)

current-version:
	@echo "Current version: $(LATEST_TAG)"

release: check-git-clean
	@echo "Please choose: make release [major|minor|patch]"
	@exit 1

major: check-git-clean
	$(eval NEW_VERSION := v$(shell echo $$(($(MAJOR) + 1))).0.0)
	@echo "Bumping major version from $(LATEST_TAG) to $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "Successfully released $(NEW_VERSION)"

minor: check-git-clean
	$(eval NEW_VERSION := v$(MAJOR).$(shell echo $$(($(MINOR) + 1))).0)
	@echo "Bumping minor version from $(LATEST_TAG) to $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "Successfully released $(NEW_VERSION)"

patch: check-git-clean
	$(eval NEW_VERSION := v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH) + 1))))
	@echo "Bumping patch version from $(LATEST_TAG) to $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "Successfully released $(NEW_VERSION)"

# Helper target to ensure working directory is clean
check-git-clean:
	@if [ -n "$(shell git status --porcelain)" ]; then \
		echo "Error: Working directory is not clean. Please commit or stash changes first."; \
		exit 1; \
	fi
	@if [ -z "$(shell git remote get-url origin 2>/dev/null)" ]; then \
		echo "Error: No git remote 'origin' found."; \
		exit 1; \
	fi 