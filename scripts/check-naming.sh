#!/usr/bin/env bash
# Checks file naming rules. Full rules: docs/conventions/naming.md
# Usage: bash scripts/check-naming.sh   (called by `make check-naming`)
set -u

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND="$ROOT/apps/backend/internal"
WEB="$ROOT/apps/web/src/features"
fail=0

err() { echo "NAMING: $1"; fail=1; }

# ---------- Backend (Go) ----------
# Folders in internal/ that are not business modules are skipped.
is_excluded_module() {
  case "$1" in platform|shared|testutil) return 0 ;; esac
  return 1
}

if [ -d "$BACKEND" ]; then
  while IFS= read -r f; do
    rel="${f#"$ROOT"/}"
    base="$(basename "$f")"
    dir="$(dirname "$f")"
    case "$f" in */gen/*) continue ;; esac
    case "$base" in *.gen.go) continue ;; esac

    module="${f#"$BACKEND"/}"
    module="${module%%/*}"
    is_excluded_module "$module" && continue

    # Tests: must have a counterpart file; then check the prefix of the source file.
    src="$base"
    case "$base" in
      *_integration_test.go) src="${base%_integration_test.go}.go" ;;
      *_test.go)             src="${base%_test.go}.go" ;;
    esac
    if [ "$src" != "$base" ] && [ ! -f "$dir/$src" ]; then
      err "$rel: test file without counterpart $src"
    fi

    # Files at the module root: only module.go, errors.go, doc.go.
    if [ "$(dirname "$dir")" = "$BACKEND" ]; then
      case "$src" in
        module.go|errors.go|doc.go) ;;
        *) err "$rel: files at the module root may only be module.go, errors.go, doc.go" ;;
      esac
      continue
    fi

    # Files in a subfolder: prefix depends on the folder.
    sub="$(basename "$dir")"
    case "$sub" in
      domain)  case "$src" in model_*.go|errors.go|doc.go) ;; *) err "$rel: files in domain/ must start with model_" ;; esac ;;
      app)     case "$src" in service_*.go|job_*.go|doc.go) ;; *) err "$rel: files in app/ must start with service_ or job_" ;; esac ;;
      httpapi) case "$src" in handler_*.go|mapper_*.go|doc.go) ;; *) err "$rel: files in httpapi/ must start with handler_ or mapper_" ;; esac ;;
      store)   case "$src" in repo_*.go|doc.go) ;; *) err "$rel: files in store/ must start with repo_" ;; esac ;;
      *)       err "$rel: unknown subfolder ($sub). Use domain/, app/, httpapi/, store/" ;;
    esac
  done < <(find "$BACKEND" -type f -name '*.go' | sort)

  # sqlc SQL queries: must start with query_
  while IFS= read -r f; do
    base="$(basename "$f")"
    case "$f" in */gen/*) continue ;; esac
    case "$base" in query_*.sql) ;; *) err "${f#"$ROOT"/}: query files must start with query_" ;; esac
  done < <(find "$BACKEND" -type f -path '*/store/queries/*' -name '*.sql' | sort)
fi

# ---------- Frontend (React) ----------
if [ -d "$WEB" ]; then
  while IFS= read -r f; do
    rel="${f#"$ROOT"/}"
    base="$(basename "$f")"
    dir="$(dirname "$f")"
    sub="$(basename "$dir")"

    # Tests: must have a counterpart file (X.test.tsx -> X.tsx or X.ts)
    case "$base" in
      *.test.tsx|*.test.ts)
        stem="${base%.test.tsx}"; [ "$stem" = "$base" ] && stem="${base%.test.ts}"
        if [ ! -f "$dir/$stem.tsx" ] && [ ! -f "$dir/$stem.ts" ]; then
          err "$rel: test file without counterpart $stem.ts(x)"
        fi
        continue
        ;;
    esac

    case "$sub" in
      pages)      case "$base" in [[:upper:]]*Page.tsx) ;; *) err "$rel: files in pages/ must be PascalCase ending in Page.tsx" ;; esac ;;
      components) case "$base" in [[:upper:]]*.tsx) ;; *) err "$rel: files in components/ must be PascalCase .tsx" ;; esac ;;
      hooks)      case "$base" in use[[:upper:]]*.ts|use[[:upper:]]*.tsx) ;; *) err "$rel: files in hooks/ must start with use" ;; esac ;;
      schemas)    case "$base" in *Schema.ts) ;; *) err "$rel: files in schemas/ must end in Schema.ts" ;; esac ;;
    esac
  done < <(find "$WEB" -type f \( -name '*.ts' -o -name '*.tsx' \) \
             \( -path '*/pages/*' -o -path '*/components/*' -o -path '*/hooks/*' -o -path '*/schemas/*' \) | sort)
fi

if [ "$fail" -ne 0 ]; then
  echo "Naming check failed. See docs/conventions/naming.md"
  exit 1
fi
echo "Naming OK"
