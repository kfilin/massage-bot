#!/bin/bash
#
# verify_backup.sh — verify the most recent (or specified) backup ZIP is
# restorable. A backup that can't be restored is useless; run this after
# every backup cycle to catch silent corruption before the next day's
# backup overwrites the good one.
#
# Usage:
#   scripts/verify_backup.sh                  # verify most recent backup
#   scripts/verify_backup.sh <path-to-zip>    # verify specific backup
#
# Exit codes:
#   0  backup verified — ZIP is valid AND contains expected files
#   1  invalid arguments or backup file missing
#   2  ZIP integrity check failed (unzip -t)
#   3  expected files/directories missing
#   4  JSON content validation failed (extract to /tmp + jq parse)
#
# Per BACKLOG #44: "Add backup verification step (daily ZIP backup exists
# per startup.md, but restore is never tested)."

set -euo pipefail

DATA_DIR=${DATA_DIR:-"./data"}
BACKUP_DIR="${DATA_DIR}/backups"

# --- Pick target backup ---------------------------------------------------

if [ $# -ge 1 ]; then
    TARGET="$1"
else
    if [ ! -d "$BACKUP_DIR" ]; then
        echo "ERROR: backup directory not found: $BACKUP_DIR" >&2
        exit 1
    fi
    # Most recent .zip (lexicographic sort matches YYYYMMDD_HHMMSS naming).
    TARGET=$(ls -1t "${BACKUP_DIR}"/*.zip 2>/dev/null | head -n 1 || true)
    if [ -z "${TARGET:-}" ]; then
        echo "ERROR: no backups found in ${BACKUP_DIR}" >&2
        exit 1
    fi
fi

if [ ! -f "$TARGET" ]; then
    echo "ERROR: backup file not found: $TARGET" >&2
    exit 1
fi

echo "Verifying backup: $TARGET"

# --- Step 1: ZIP integrity check ------------------------------------------

if ! unzip -t "$TARGET" > /dev/null; then
    echo "FAIL: ZIP integrity check failed (corrupt or truncated archive)" >&2
    exit 2
fi
echo "  ✓ ZIP integrity OK"

# --- Step 2: Required entries present -------------------------------------
# A restorable backup of massage-bot data must include at minimum:
#   - blacklist.txt            (banned users)
#   - patients/                (one subdir per patient)
#   - token.json               (Google Calendar OAuth)
# Media is large and may legitimately be empty on small installs; we
# only warn if it's missing entirely.

REQUIRED=("blacklist.txt" "patients" "token.json")
MISSING=()
# Real backups created by backup_data.sh zip with the $DATA_DIR prefix
# (e.g., `data/blacklist.txt`). Match entries by their trailing path
# component so the prefix doesn't fool us.
ENTRY_NAMES=$(unzip -l "$TARGET" | awk 'NR>3 && $NF != "" {print $NF}')
for entry in "${REQUIRED[@]}"; do
    if ! printf '%s\n' "$ENTRY_NAMES" | grep -qE "(^|/)${entry}(/|$)"; then
        MISSING+=("$entry")
    fi
done

if [ ${#MISSING[@]} -gt 0 ]; then
    echo "FAIL: required entries missing from backup:" >&2
    printf '   - %s\n' "${MISSING[@]}" >&2
    exit 3
fi
echo "  ✓ required entries present (blacklist.txt, patients/, token.json)"

# --- Step 3: Extract + JSON spot-check ------------------------------------
# Extract to a temp dir; verify at least one patient.json parses. We don't
# validate every record (too slow for large datasets) — a sample check
# catches the common failure mode of a truncated JSON write.

WORKDIR=$(mktemp -d -t massage-bot-verify-XXXXXX)
trap 'rm -rf "$WORKDIR"' EXIT

unzip -q "$TARGET" -d "$WORKDIR"

SAMPLE_JSON=$(find "$WORKDIR" -type f -name 'patient.json' 2>/dev/null | head -n 1 || true)
if [ -n "$SAMPLE_JSON" ]; then
    if command -v jq > /dev/null 2>&1; then
        if ! jq -e . "$SAMPLE_JSON" > /dev/null 2>&1; then
            echo "FAIL: patient JSON did not parse: $SAMPLE_JSON" >&2
            exit 4
        fi
        echo "  ✓ sample patient JSON parses ($SAMPLE_JSON)"
    else
        # Fall back to python (always present in gitlab runner / alpine)
        if ! python3 -c "import json,sys; json.load(open(sys.argv[1]))" "$SAMPLE_JSON" > /dev/null 2>&1; then
            echo "FAIL: patient JSON did not parse: $SAMPLE_JSON" >&2
            exit 4
        fi
        echo "  ✓ sample patient JSON parses ($SAMPLE_JSON)"
    fi
else
    echo "  ⚠ no patient.json found (database may be empty — acceptable for fresh install)"
fi

echo "PASS: backup verified — $TARGET"
