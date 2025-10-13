import os

FOLDER = "."
TEMP_NAME = "__test_build.exe"

files = os.scandir(FOLDER)
files = filter(lambda x: x.name.endswith(".c"), files)
files = map(lambda x: x.name, files)
files = list(files)


os.system(f"gcc -o {TEMP_NAME} {" ".join(files)}")

try:
    os.system(TEMP_NAME)
finally:
    os.unlink(TEMP_NAME)
