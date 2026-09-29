"""Validate a gcsim submission archive and prepare a non-destructive D1 import.

This command only writes a SQL file. Apply it with Wrangler after a verified backup.
"""
import argparse
import hashlib
import json
import math
from pathlib import Path, PurePosixPath
import re
import tarfile
import time

ID = re.compile(r"^[a-zA-Z0-9_-]{1,128}$")

def quoted(value):
    return "'" + str(value).replace("'", "''") + "'"

def load_entries(archive):
    entries, seen = [], set()
    with tarfile.open(archive, "r:xz") as source:
        for item in source:
            path = PurePosixPath(item.name)
            if path.is_absolute() or ".." in path.parts or item.issym() or item.islnk():
                raise ValueError("Unsafe archive member")
            if item.isdir():
                continue
            if not item.isfile() or len(path.parts) != 3 or path.parts[0] != "gcsim-db-submissions" or path.parts[1] not in ("approved", "not approved") or path.suffix != ".json" or item.size > 1_000_000:
                raise ValueError("Unexpected archive member")
            entry = json.load(source.extractfile(item))
            record_id = entry.get("_id", "")
            if not ID.fullmatch(record_id) or not ID.fullmatch(entry.get("share_key", "")) or record_id in seen:
                raise ValueError("Invalid or duplicate record ID")
            approved = path.parts[1] == "approved"
            if approved != bool(entry.get("is_db_valid")):
                raise ValueError("Approval state does not match archive folder")
            if not isinstance(entry.get("config"), str) or not entry["config"].strip():
                raise ValueError("Missing simulation config")
            for number in (entry.get("create_date"), entry.get("last_update"), entry.get("summary", {}).get("mean_dps_per_target"), entry.get("summary", {}).get("sim_duration", {}).get("mean")):
                if not isinstance(number, (int, float)) or not math.isfinite(number) or number < 0:
                    raise ValueError("Invalid timestamp or simulation summary")
            if not isinstance(entry["summary"].get("team"), list) or not 1 <= len(entry["summary"]["team"]) <= 4:
                raise ValueError("Invalid team")
            # Reject NaN and infinity anywhere in the document.
            json.dumps(entry, allow_nan=False)
            seen.add(record_id)
            entries.append((approved, entry))
    if not entries:
        raise ValueError("The archive is empty")
    return entries

def make_sql(entries, archive_hash, now):
    statements = ["-- KQM archive import. Existing local records and review decisions are preserved."]
    for approved, entry in entries:
        document = json.dumps(entry, ensure_ascii=False, separators=(",", ":"), allow_nan=False)
        doc = quoted(document)
        record_id = quoted(entry["_id"])
        if approved:
            statements.append(f"""INSERT INTO simulations(id,document,create_date,dps,duration,imported_at,seen_run,visible,source)
VALUES({record_id},{doc},{int(entry['create_date'])},{entry['summary']['mean_dps_per_target']},{entry['summary']['sim_duration']['mean']},{now},{quoted(archive_hash)},1,'archive')
ON CONFLICT(id) DO UPDATE SET document=excluded.document,create_date=excluded.create_date,dps=excluded.dps,duration=excluded.duration,imported_at=excluded.imported_at,seen_run=excluded.seen_run,visible=1,source='archive'
WHERE simulations.source='upstream' AND COALESCE(json_extract(simulations.document,'$.last_update'),0) <= COALESCE(json_extract(excluded.document,'$.last_update'),0);""")
        else:
            digest = hashlib.sha256(document.encode()).hexdigest()
            statements.append(f"""INSERT INTO submissions(id,request_hash,source_url,document,status,submitted_at)
VALUES({record_id},{quoted(digest)},{quoted('https://taghelper.simpact.app/id/' + entry['_id'])},{doc},'pending',{int(entry['create_date']) * 1000}) ON CONFLICT(id) DO NOTHING;""")
    return "\n".join(statements) + "\n"

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("archive", type=Path)
    parser.add_argument("--out", required=True, type=Path)
    args = parser.parse_args()
    entries = load_entries(args.archive)
    digest = hashlib.sha256(args.archive.read_bytes()).hexdigest()
    sql = make_sql(entries, digest, int(time.time() * 1000))
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(sql)
    args.out.chmod(0o600)
    print(json.dumps({"approved": sum(approved for approved, _ in entries), "pending": sum(not approved for approved, _ in entries), "archive_sha256": digest, "sql": str(args.out)}))

if __name__ == "__main__":
    main()
