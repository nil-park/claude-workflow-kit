FG_BOLD := \033[1m
FG_WHITE := \033[97m
BG_GREEN := \033[42m
BG_BLUE := \033[44m
RESET := \033[0m

.PHONY: format test

# Markdown / JSON / YAML 포매팅 후 매니페스트 검증까지 이어서 수행한다.
# 확장자별 파일이 없어도 오류 없이 넘어간다.
format:
	@echo -e "\n$(BG_BLUE)$(FG_WHITE)$(FG_BOLD) prettier $(RESET)\n"
	@npx --yes prettier --write --no-error-on-unmatched-pattern "**/*.md" "**/*.json" "**/*.yml" "**/*.yaml"
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) claude plugin validate $(RESET)\n"
	@claude plugin validate .

# 마켓플레이스 / 플러그인 매니페스트 검증.
test:
	@echo -e "\n$(BG_GREEN)$(FG_WHITE)$(FG_BOLD) claude plugin validate $(RESET)\n"
	@claude plugin validate .
