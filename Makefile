FG_BOLD := \033[1m
FG_BLACK := \033[30m
FG_WHITE := \033[97m
BG_GREEN := \033[42m
BG_YELLOW := \033[43m
BG_BLUE := \033[44m
BG_PINK := \033[45m
RESET := \033[0m

.PHONY: format test statusline duo-relay agy-compaction-hook claude-queue

GO_MODULES := statusline duo-relay agy-compaction-hook claude-queue

# 고치면서 검사한다. 확장자별 파일이 없어도 오류 없이 넘어간다.
format:
	@echo -e "\n$(BG_BLUE)$(FG_WHITE)$(FG_BOLD) prettier $(RESET)\n"
	@npx --yes prettier --write --no-error-on-unmatched-pattern "**/*.md" "**/*.json" "**/*.yml" "**/*.yaml"
	@echo -e "\n$(BG_BLUE)$(FG_WHITE)$(FG_BOLD) ruff check --fix $(RESET)"
	@uv run ruff check --fix
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) ruff format $(RESET)"
	@uv run ruff format
	@echo -e "\n$(BG_YELLOW)$(FG_BLACK)$(FG_BOLD) pyright $(RESET)\n"
	@uv run pyright
	@echo -e "\n$(BG_PINK)$(FG_WHITE)$(FG_BOLD) pytest $(RESET)\n"
	@uv run pytest
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) gofmt -w $(RESET)\n"
	@gofmt -w $(GO_MODULES)
	@echo -e "\n$(BG_YELLOW)$(FG_BLACK)$(FG_BOLD) go vet $(RESET)\n"
	@for m in $(GO_MODULES); do (cd $$m && go vet ./...) || exit 1; done
	@echo -e "\n$(BG_PINK)$(FG_WHITE)$(FG_BOLD) go test $(RESET)\n"
	@for m in $(GO_MODULES); do (cd $$m && go test ./...) || exit 1; done
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) claude plugin validate $(RESET)\n"
	@claude plugin validate .

# 고치지 않고 검사만 한다.
test:
	@echo -e "\n$(BG_BLUE)$(FG_WHITE)$(FG_BOLD) ruff check $(RESET)"
	@uv run ruff check
	@echo -e "\n$(BG_YELLOW)$(FG_BLACK)$(FG_BOLD) pyright $(RESET)\n"
	@uv run pyright
	@echo -e "\n$(BG_PINK)$(FG_WHITE)$(FG_BOLD) pytest $(RESET)\n"
	@uv run pytest
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) gofmt -l $(RESET)\n"
	@unformatted="$$(gofmt -l $(GO_MODULES))"; if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi
	@echo -e "\n$(BG_YELLOW)$(FG_BLACK)$(FG_BOLD) go vet $(RESET)\n"
	@for m in $(GO_MODULES); do (cd $$m && go vet ./...) || exit 1; done
	@echo -e "\n$(BG_PINK)$(FG_WHITE)$(FG_BOLD) go test $(RESET)\n"
	@for m in $(GO_MODULES); do (cd $$m && go test ./...) || exit 1; done
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) claude plugin validate $(RESET)\n"
	@claude plugin validate .

# build/statusline을 빌드한다. Windows에서는 go env GOEXE가 .exe 확장자를 붙인다.
statusline:
	@cd statusline && go build -o "../build/statusline$$(go env GOEXE)" .

# duo-relay를 ~/.local/bin에 설치한다.
duo-relay:
	@cd duo-relay && go build -o "$(HOME)/.local/bin/duo-relay$$(go env GOEXE)" .

# agy-compaction-hook을 ~/.local/bin에 설치한다.
agy-compaction-hook:
	@cd agy-compaction-hook && go build -o "$(HOME)/.local/bin/agy-compaction-hook$$(go env GOEXE)" .

# claude-queue를 ~/.local/bin에 설치한다.
claude-queue:
	@cd claude-queue && go build -o "$(HOME)/.local/bin/claude-queue$$(go env GOEXE)" .
