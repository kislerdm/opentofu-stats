REPO := opentofu/opentofu
SQLLITE_STORE := $(PWD)/data/tofu.db
JSON_OUT := $(PWD)/data/aggregates.json

.PHONY: extract
extract: ## Extracts the stats data.
	@ github-to-sqlite commits $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite issues $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite pull-requests $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite stargazers $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite contributors $(SQLLITE_STORE) $(REPO)
	@ go run cmd/extract-downloads/main.go -db $(SQLLITE_STORE)

.PHONY: transform
transform: ## Transforms data.
	@ go run cmd/transform/main.go -db $(SQLLITE_STORE) -out $(JSON_OUT)
