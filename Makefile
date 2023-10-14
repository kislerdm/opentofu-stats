REPO := opentofu/opentofu
SQLLITE_STORE := tofu-stats.db

.PHONY: el
el: # Extracts and loads stats data.
	@ github-to-sqlite commits $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite issues $(SQLLITE_STORE) $(REPO) && \
		github-to-sqlite pull-requests $(SQLLITE_STORE) $(REPO)
