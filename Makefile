.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

REPO := opentofu/opentofu
SQLLITE_STORE := $(PWD)/data/tofu.db
JSON_OUT := $(PWD)/data/aggregates.json

.PHONY: downloads
downloads:
	@ go run cmd/extract-downloads/main.go -db $(SQLLITE_STORE)

.PHONY: stats
stats:
	@ github-to-sqlite commits $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite issues $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite pull-requests $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite stargazers $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite contributors $(SQLLITE_STORE) $(REPO)


.PHONY: extract
extract: stats downloads ## Extracts the stats data for opentofu/opentofu.

.PHONY: transform
transform: ## Transforms data for opentofu/opentofu.
	@ go run cmd/transform/main.go -db $(SQLLITE_STORE) -out $(JSON_OUT)

SQLLITE_STORE_TF := $(PWD)/data/tf.db
JSON_OUT_TF := $(PWD)/data/aggregates-tf.json

.PHONY: stats-tf
stats-tf:
	@ github-to-sqlite commits $(SQLLITE_STORE_TF) hashicorp/terraform && \
		github-to-sqlite issues $(SQLLITE_STORE_TF) hashicorp/terraform && \
		github-to-sqlite pull-requests $(SQLLITE_STORE_TF) hashicorp/terraform && \
		github-to-sqlite stargazers $(SQLLITE_STORE_TF) hashicorp/terraform && \
		github-to-sqlite contributors $(SQLLITE_STORE_TF) hashicorp/terraform

.PHONY: downloads-tf
downloads-tf:
	@ go run cmd/extract-downloads/main.go -db $(SQLLITE_STORE_TF) -owner hashicorp -name terraform

.PHONY: extract-tf
extract-tf: stats-tf downloads-tf ## Extracts the stats data for hashicorp/terraform.

.PHONY: transform-tf
transform-tf: ## Transforms data for hashicorp/terraform.
	@ go run cmd/transform/main.go -db $(SQLLITE_STORE_TF) -out $(JSON_OUT_TF)

.NOTPARALLEL:
