#!/usr/bin/env bash
set -euo pipefail
script_directory=$(dirname "$0")
cd "$script_directory/.."
root=$PWD
shfmt=$(cd tools && go tool -n shfmt)
if [[ -x bin/shellcheck ]]; then
  checker=$root/bin/shellcheck
else
  checker=$(command -v shellcheck) || {
    printf 'ShellCheck is missing; run bash scripts/setup-shellcheck.sh.\n' >&2
    exit 1
  }
fi
expected=$(cat .shellcheck-version)
installed=$("$checker" --version)
if [[ "$installed" != *"version: $expected"$'\n'* && "$installed" != *"version: $expected" ]]; then
  printf 'ShellCheck %s is required; run bash scripts/setup-shellcheck.sh.\n' "$expected" >&2
  exit 1
fi

check_scripts() {
  local script
  for script in "$@"; do
    bash -n "$script" || return 1
  done
  "$checker" --rcfile="$root/.shellcheckrc" --shell=bash --severity=style "$@" || return 1
  "$shfmt" -d -ln bash -i 2 -ci "$@" || return 1
}

temporary=$(mktemp -d)
trap 'rm -rf "$temporary"' EXIT
if [[ "${1:-}" == --self-test ]]; then
  printf '#!/usr/bin/env bash\nprintf "%%s\\n" "hello"\n' >"$temporary/good.sh"
  check_scripts "$temporary/good.sh"
  printf '#!/usr/bin/env bash\nif then\n' >"$temporary/syntax.sh"
  # Keep the expansion in the generated negative fixture.
  # shellcheck disable=SC2016
  printf '#!/usr/bin/env bash\nvalue="two words"\nprintf "%%s\\n" $value\n' >"$temporary/lint.sh"
  printf '#!/usr/bin/env bash\nif true;then echo ok;fi\n' >"$temporary/format.sh"
  for probe in syntax lint format; do
    # check_scripts returns explicitly after each failed command.
    # shellcheck disable=SC2310
    if check_scripts "$temporary/$probe.sh" >"$temporary/$probe.log" 2>&1; then
      printf 'Shell gate accepted the %s violation.\n' "$probe" >&2
      exit 1
    fi
  done
  # Confirm each probe reached the intended check, rather than failing incidentally.
  grep -q 'syntax error' "$temporary/syntax.log"
  grep -q 'SC2086' "$temporary/lint.log"
  grep -q '^@@ ' "$temporary/format.log"
  printf 'Shell gate rejected syntax, quoting, and formatting violations.\n'
  exit 0
fi
if [[ $# != 0 ]]; then
  printf 'Usage: bash scripts/lint-shell.sh [--self-test]\n' >&2
  exit 1
fi

# Include new scripts before they are staged; honor Git ignores and retain spaces.
git ls-files --cached --others --exclude-standard -z >"$temporary/files"
scripts=()
while IFS= read -r -d '' file; do
  [[ -f "$file" ]] || continue
  case "$file" in
    *.sh | *.bash) scripts+=("$file") ;;
    *)
      first_line=
      IFS= read -r first_line <"$file" || true
      if [[ "$first_line" == '#!'* && "$first_line" =~ (^|[[:space:]/])bash([[:space:]]|$) ]]; then
        scripts+=("$file")
      fi
      ;;
  esac
done <"$temporary/files"
if [[ ${#scripts[@]} == 0 ]]; then
  printf 'No owned Bash scripts were discovered.\n' >&2
  exit 1
fi
check_scripts "${scripts[@]}"
printf 'Checked %s Bash scripts.\n' "${#scripts[@]}"
