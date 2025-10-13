import os
import subprocess
import pathlib
import re

TEMP_NAME = "__test_build.exe"
TEST_FOLDER = "tests"
EXPECTED_MARK = "// expect:"

os.system(f"go build -o {TEMP_NAME} .")


def read_expected(file: pathlib.Path) -> list[str]:
    with open(file, "r", encoding="utf-8") as f:
        expected = [
            re.search(f"{EXPECTED_MARK}(.+)", line).group(1).strip()
            for line in f.readlines()
            if EXPECTED_MARK in line
        ]
    return expected


def run_script(file: pathlib.Path) -> tuple[str, str]:
    result = subprocess.run(
        [TEMP_NAME, file],
        capture_output=True,
        text=True,
    )
    return result.stdout, result.stderr


def check_out(out: str, expected: list[str]) -> str:
    out = out.splitlines()
    if out == expected:
        return "OK"
    if len(out) != len(expected):
        return f"want {len(expected)} lines, got {len(out)}"
    for i, line in enumerate(out):
        if line != expected[i]:
            return f"expected '{expected[i]}', got '{line}'"


def short_string(string: str, length: int) -> str:
    if len(string) > length:
        return string[:length]
    return string


try:
    ok_flag = True
    for file in pathlib.Path(TEST_FOLDER).rglob("*.ndl"):
        expected = read_expected(file)
        out, err = run_script(file)

        if err != "":
            ok_flag = False
            print(file, "-> error: stderr:", short_string(err, 256).strip())
            continue

        result = check_out(out, expected)
        if result == "OK":
            print(file, "-> ok")
        else:
            ok_flag = False
            print(file, "-> error:", result)

    print("=" * 8, "result", "=" * 8)
    print("OK!" if ok_flag else "ERROR!")
finally:
    os.unlink(TEMP_NAME)
