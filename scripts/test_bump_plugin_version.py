import json
import tempfile
import unittest
from pathlib import Path

from scripts.bump_plugin_version import bump_file_version, bump_version


class BumpVersionTest(unittest.TestCase):
    def test_bumps_patch_version(self) -> None:
        self.assertEqual(bump_version("0.0.1"), "0.0.2")

    def test_rolls_patch_into_minor(self) -> None:
        self.assertEqual(bump_version("0.0.99"), "0.1.0")

    def test_rejects_minor_overflow(self) -> None:
        with self.assertRaises(ValueError):
            bump_version("0.99.99")

    def test_updates_plugin_json_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            path = Path(tmpdir) / "plugin.json"
            path.write_text(
                json.dumps(
                    {
                        "name": "go-arch-guard",
                        "description": "Adds the go-arch-guard skill.",
                        "version": "0.0.99",
                    }
                ),
                encoding="utf-8",
            )

            next_version = bump_file_version(path)

            self.assertEqual(next_version, "0.1.0")
            data = json.loads(path.read_text(encoding="utf-8"))
            self.assertEqual(data["version"], "0.1.0")


if __name__ == "__main__":
    unittest.main()
