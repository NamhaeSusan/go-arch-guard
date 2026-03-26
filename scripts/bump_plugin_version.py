import json
import sys
from pathlib import Path


def parse_version(version: str) -> tuple[int, int, int]:
    parts = version.split(".")
    if len(parts) != 3:
        raise ValueError(f"invalid version: {version}")

    values = []
    for part in parts:
        if not part.isdigit():
            raise ValueError(f"invalid version: {version}")
        value = int(part)
        if value < 0 or value > 99:
            raise ValueError(f"version component out of range: {version}")
        values.append(value)
    return values[0], values[1], values[2]


def bump_version(version: str) -> str:
    major, minor, patch = parse_version(version)
    patch += 1
    if patch <= 99:
        return f"{major}.{minor}.{patch}"

    minor += 1
    if minor > 99:
        raise ValueError(f"version overflow: {version}")
    return f"{major}.{minor}.0"


def bump_file_version(path: Path) -> str:
    data = json.loads(path.read_text(encoding="utf-8"))
    next_version = bump_version(data["version"])
    data["version"] = next_version
    path.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    return next_version


def main(argv: list[str]) -> int:
    path = Path(argv[1]) if len(argv) > 1 else Path(
        "plugins/go-arch-guard/.claude-plugin/plugin.json"
    )
    next_version = bump_file_version(path)
    print(next_version)
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
