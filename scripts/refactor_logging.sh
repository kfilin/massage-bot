#!/bin/bash
# Phase 2: Automated Logging Consolidation Script
# Replaces log.Printf calls with logging.* equivalents

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🔧 Phase 2: Logging Consolidation Automation"
echo "============================================="
echo ""

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Counter for changes
TOTAL_REPLACEMENTS=0

# Function to process a single file
process_file() {
    local file="$1"
    local changes=0
    
    echo -e "${YELLOW}Processing:${NC} $file"
    
    # Check if file imports "log" package
    if ! grep -q '^import "log"' "$file" && ! grep -q '"log"' "$file"; then
        echo "  ⏭️  Skipping (no log import)"
        return 0
    fi
    
    # Create backup
    cp "$file" "$file.bak"
    
    # Replace log.Printf with appropriate logging level based on content
    # DEBUG patterns
    if grep -q 'log.Printf("DEBUG' "$file"; then
        sed -i 's/log\.Printf("DEBUG\([^"]*\)"/logging.Debugf("\1"/g' "$file"
        changes=$((changes + $(grep -c 'logging.Debugf' "$file" || echo 0)))
    fi
    
    # ERROR patterns
    if grep -q 'log.Printf("ERROR' "$file" || grep -q 'log.Printf(".*Error:' "$file"; then
        sed -i 's/log\.Printf("ERROR\([^"]*\)"/logging.Errorf("\1"/g' "$file"
        sed -i 's/log\.Printf("\([^"]*\)Error:/logging.Errorf("\1Error:/g' "$file"
        changes=$((changes + $(grep -c 'logging.Errorf' "$file" || echo 0)))
    fi
    
    # WARNING patterns
    if grep -q 'log.Printf("Warning:' "$file" || grep -q 'log.Printf("WARN' "$file"; then
        sed -i 's/log\.Printf("Warning:/logging.Warnf("Warning:/g' "$file"
        sed -i 's/log\.Printf("WARN\([^"]*\)"/logging.Warnf("\1"/g' "$file"
        changes=$((changes + $(grep -c 'logging.Warnf' "$file" || echo 0)))
    fi
    
    # AUTH ERROR patterns (use Errorf)
    if grep -q 'log.Printf("AUTH ERROR:' "$file"; then
        sed -i 's/log\.Printf("AUTH ERROR:/logging.Errorf("AUTH ERROR:/g' "$file"
    fi
    
    # Cancel Error patterns
    if grep -q 'log.Printf("Cancel Error:' "$file"; then
        sed -i 's/log\.Printf("Cancel Error:/logging.Errorf("Cancel Error:/g' "$file"
    fi
    
    # WebDAV patterns (use Infof for normal operations, Warnf for auth issues)
    if grep -q 'log.Printf("WebDAV \[Err\]' "$file"; then
        sed -i 's/log\.Printf("WebDAV \[Err\]/logging.Errorf("WebDAV [Err]/g' "$file"
    fi
    if grep -q 'log.Printf("WebDAV \[Auth' "$file"; then
        sed -i 's/log\.Printf("WebDAV \[Auth/logging.Warnf("WebDAV [Auth]/g' "$file"
    fi
    if grep -q 'log.Printf("WebDAV' "$file"; then
        sed -i 's/log\.Printf("WebDAV/logging.Infof("WebDAV/g' "$file"
    fi
    
    # DB Error patterns
    if grep -q 'log.Printf("DB Error' "$file"; then
        sed -i 's/log\.Printf("DB Error/logging.Errorf("DB Error/g' "$file"
    fi
    
    # Remaining log.Printf (default to Infof for informational messages)
    if grep -q 'log.Printf(' "$file"; then
        sed -i 's/log\.Printf(/logging.Infof(/g' "$file"
        changes=$((changes + $(grep -c 'logging.Infof' "$file" || echo 0)))
    fi
    
    # Add logging import if not present
    if ! grep -q '"github.com/kfilin/massage-bot/internal/logging"' "$file"; then
        # Find the import block and add logging import
        if grep -q '^import (' "$file"; then
            # Multi-line import
            sed -i '/^import (/a\	"github.com/kfilin/massage-bot/internal/logging"' "$file"
        elif grep -q '^import "' "$file"; then
            # Single import, convert to multi-line
            sed -i 's/^import "\([^"]*\)"/import (\n	"\1"\n	"github.com\/kfilin\/massage-bot\/internal\/logging"\n)/' "$file"
        fi
        echo "  ✅ Added logging import"
    fi
    
    # Remove "log" import if no longer used
    if ! grep -q '\blog\.' "$file"; then
        sed -i '/^[[:space:]]*"log"$/d' "$file"
        # Clean up empty import blocks
        sed -i '/^import ($/,/^)$/{/^import ($/N;/^import (\n)$/d;}' "$file"
        echo "  ✅ Removed unused log import"
    fi
    
    # Check if file actually changed
    if ! diff -q "$file" "$file.bak" > /dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} Modified ($changes replacements)"
        TOTAL_REPLACEMENTS=$((TOTAL_REPLACEMENTS + changes))
        rm "$file.bak"
    else
        echo "  ⏭️  No changes needed"
        mv "$file.bak" "$file"
    fi
}

# Find all Go files with log.Printf
echo "🔍 Finding files with log.Printf..."
FILES=$(grep -rl "log\.Printf" --include="*.go" . | grep -v ".bak" | grep -v "vendor/" || true)

if [ -z "$FILES" ]; then
    echo -e "${GREEN}✅ No files found with log.Printf - Phase 2 already complete!${NC}"
    exit 0
fi

echo "Found $(echo "$FILES" | wc -l) files to process"
echo ""

# Process each file
for file in $FILES; do
    process_file "$file"
done

echo ""
echo "============================================="
echo -e "${GREEN}✅ Phase 2 Complete!${NC}"
echo "Total replacements: $TOTAL_REPLACEMENTS"
echo ""
echo "Next steps:"
echo "1. Review changes: git diff"
echo "2. Run tests: make test"
echo "3. Build: go build ./..."
echo "4. Commit: git commit -am 'refactor: consolidate logging (Phase 2)'"
