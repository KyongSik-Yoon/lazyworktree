#!/usr/bin/env -S uv --quiet run --script
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "textual",
#     "rich",
#     "PyYAML",
# ]
# ///

import sys
from lazyworktree.app import GitWtStatus

def main():
    initial_filter = " ".join(sys.argv[1:]).strip()
    app = GitWtStatus(initial_filter=initial_filter)
    run_result = app.run()
    if run_result is None:
        sys.exit(0)
    result = run_result
    if result:
        print(result)

if __name__ == "__main__":
    main()
