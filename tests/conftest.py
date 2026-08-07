import sys
from pathlib import Path

# 훅은 남의 시스템 python3로 직접 실행되므로 패키지가 아니다. 경로로 붙인다.
HOOKS = Path(__file__).resolve().parents[1] / "plugins" / "ko-style" / "hooks"
sys.path.insert(0, str(HOOKS))
