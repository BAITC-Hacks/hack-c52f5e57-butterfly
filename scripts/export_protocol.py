"""Local export subprocess used by the Go HTTP server; no network access."""
import json
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))
from butterfly.exports import render_export

if __name__ == "__main__":
    meeting = json.load(sys.stdin)
    sys.stdout.buffer.write(render_export(meeting, sys.argv[1]))
