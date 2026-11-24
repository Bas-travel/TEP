#!/usr/bin/env python3
"""Fail when the Go route table and docs/api/openapi.yaml disagree (every route documented, every operation served)."""
import glob, re, sys
import yaml

spec = yaml.safe_load(open("docs/api/openapi.yaml"))
ROOTED = ("/health", "/partner/v1", "/admin/v1", "/staff/v1")
doc = set()
for path, item in spec["paths"].items():
    full = path if path.startswith(ROOTED) else "/v1" + path
    for m in item:
        if m in ("get", "post", "put", "patch", "delete"):
            doc.add((m.upper(), re.sub(r"\{[^}]+\}", "{}", full)))
code = set()
for f in glob.glob("services/api-go/internal/api/*.go"):
    if f.endswith("_test.go"):
        continue
    for m, p in re.findall(r'HandleFunc\("([A-Z]+) (/[^"]+)"', open(f).read()):
        code.add((m, re.sub(r"\{[^}]+\}", "{}", p)))
missing, extra = sorted(code - doc), sorted(doc - code)
print(f"routes in code: {len(code)}, operations in spec: {len(doc)}")
for m, p in missing: print("NOT DOCUMENTED:", m, p)
for m, p in extra: print("NOT IMPLEMENTED:", m, p)
sys.exit(1 if missing or extra else 0)
