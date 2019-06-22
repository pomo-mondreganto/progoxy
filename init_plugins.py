import os
import shutil
import sys

FORCE = ('--force' in sys.argv)

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

RESULT_DIR = os.path.join(BASE_DIR, 'resources', 'plugins')
if not os.path.isdir(RESULT_DIR):
    os.makedirs(RESULT_DIR)

PLUGINS_DIR = os.path.join(BASE_DIR, 'plugins')

for filename in os.listdir(PLUGINS_DIR):
    if filename[0] == '.':
        continue
    dirname = os.path.join(PLUGINS_DIR, filename)
    if not os.path.isdir(dirname):
        continue

    so_filename = os.path.join(dirname, filename + '.so')
    so_result = os.path.join(RESULT_DIR, filename + '.so')

    if os.path.exists(so_result):
        if not FORCE:
            continue
        else:
            os.remove(so_result)

    os.chdir(dirname)
    os.system("go build -buildmode=\"plugin\" .")

    shutil.move(so_filename, RESULT_DIR)
