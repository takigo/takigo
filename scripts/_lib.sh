#!/usr/bin/env bash
# scripts/_lib.sh -- shared helpers for the demo-comparison scripts.
#
# Source this from any script that needs:
#   - PROJECT_DIR, SCRIPT_DIR, SS_DIR, LOGS_DIR, BIN_DIR path variables
#   - tcl_demo_for <go_demo>           — map a Go demo name to its Tcl counterpart
#   - run_compare <go_demo> [tcl_demo] — run demo_compare.sh and print just the score
#   - set_skip_if_exists <0|1>         — export SKIP_IF_EXISTS for screenshot reuse
#   - odiff_score <base> <cmp> [diff]  — run odiff, print diff % (0-100)
#   - find_windows_exact <title>       — list window IDs by exact WM_NAME
#   - maybe_xvfb "$@"                  — re-exec under xvfb-run when headless
#   - demotitle_bin                    — cached path to the built demotitle CLI
#   - demodiff_bin                     — cached path to the built demodiff CLI
#   - pin_fonts                        — private DejaVu-only fontconfig
#
# Not executable on its own; sourced by other scripts.

# Resolve the absolute paths from this file's location. PROJECT_DIR is the
# takigo repo root; SCRIPT_DIR is the directory containing this _lib.sh.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SS_DIR="$PROJECT_DIR/tmp/screenshots"
LOGS_DIR="$PROJECT_DIR/tmp/logs"
BIN_DIR="$PROJECT_DIR/tmp/bin"
# tmp/ is a Go module of its own so that scratch programs under it never
# break `go build ./...` (scripts/check_tutorial.sh writes the same file).
[[ -f "$PROJECT_DIR/tmp/go.mod" ]] || { mkdir -p "$PROJECT_DIR/tmp"; printf 'module takigo-tmp\n\ngo 1.27.0\n' > "$PROJECT_DIR/tmp/go.mod"; }
mkdir -p "$SS_DIR" "$LOGS_DIR" "$BIN_DIR"

# maybe_xvfb [ARGS...]
# Re-exec the calling script under xvfb-run when there is no usable display.
# Call it near the top of a script, before any DISPLAY-dependent work, passing
# the script's own arguments:  maybe_xvfb "$@"
# Force headless with HEADLESS=1; otherwise auto-detect an unset/unconnectable
# DISPLAY. The _XVFB_RUNNING guard prevents infinite re-exec. Xvfb runs at a
# fixed 4000x3000x24 / 96dpi so captures are deterministic; the large screen
# keeps the pointer (stuck at the screen centre, xdotool cannot move it under
# Xvfb) outside demo windows placed at +300+300, so nothing starts hovered.
maybe_xvfb() {
    [[ -n "${_XVFB_RUNNING:-}" ]] && return 0
    local need=0
    [[ "${HEADLESS:-0}" == "1" ]] && need=1
    if [[ -z "${DISPLAY:-}" ]] || ! xdpyinfo >/dev/null 2>&1; then
        need=1
    fi
    if [[ "$need" == "1" ]]; then
        command -v xvfb-run >/dev/null 2>&1 || {
            echo "HEADLESS/display unavailable but xvfb-run not found (install xorg-server-xvfb)" >&2
            exit 1
        }
        export _XVFB_RUNNING=1
        exec xvfb-run -a -s "-screen 0 4000x3000x24 -dpi 96" "$0" "$@"
    fi
}

# pin_fonts
# Points fontconfig (used by both takigo's Xft backend and Tk's Xft build) at
# a private config exposing only the DejaVu Sans/Serif/Mono families, with
# every other family aliased onto them and fixed Xft render settings. This
# removes host font-set and fallback differences from the comparison.
# Disable with PIN_FONTS=0. Idempotent; the generated tree lives in tmp/.
pin_fonts() {
    [[ "${PIN_FONTS:-1}" == "1" ]] || return 0
    [[ -n "${_TAKIGO_FONTS_PINNED:-}" ]] && return 0
    local root="$PROJECT_DIR/tmp/fontconfig"
    local conf="$root/fonts.conf"
    if [[ ! -f "$conf" ]]; then
        mkdir -p "$root/fonts" "$root/cache"
        local files
        files=$(env -u FONTCONFIG_FILE fc-list -f '%{file}\n' |
            grep -E '/DejaVu(Sans|SansMono|Serif)(-Bold|-Oblique|-BoldOblique|-Italic|-BoldItalic)?\.ttf$' | sort -u)
        if [[ -z "$files" ]]; then
            echo "pin_fonts: DejaVu fonts not found; set PIN_FONTS=0 or install ttf-dejavu" >&2
            return 1
        fi
        local f
        while read -r f; do ln -sf "$f" "$root/fonts/$(basename "$f")"; done <<<"$files"
        {
            cat <<EOF
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
<fontconfig>
  <dir>$root/fonts</dir>
  <cachedir>$root/cache</cachedir>
EOF
            local fam target
            for fam in sans-serif Helvetica Arial "Liberation Sans" "Noto Sans" system-ui; do
                printf '  <alias binding="same"><family>%s</family><prefer><family>DejaVu Sans</family></prefer></alias>\n' "$fam"
            done
            for fam in serif Times "Times New Roman" "Liberation Serif"; do
                printf '  <alias binding="same"><family>%s</family><prefer><family>DejaVu Serif</family></prefer></alias>\n' "$fam"
            done
            for fam in monospace Courier "Courier New" fixed "Liberation Mono" "Noto Sans Mono"; do
                printf '  <alias binding="same"><family>%s</family><prefer><family>DejaVu Sans Mono</family></prefer></alias>\n' "$fam"
            done
            cat <<'EOF'
  <match target="pattern">
    <edit name="family" mode="append_last"><string>DejaVu Sans</string></edit>
  </match>
  <match target="font">
    <edit name="antialias" mode="assign"><bool>true</bool></edit>
    <edit name="hinting" mode="assign"><bool>true</bool></edit>
    <edit name="hintstyle" mode="assign"><const>hintslight</const></edit>
    <edit name="autohint" mode="assign"><bool>false</bool></edit>
    <edit name="rgba" mode="assign"><const>none</const></edit>
    <edit name="lcdfilter" mode="assign"><const>lcdnone</const></edit>
    <edit name="embeddedbitmap" mode="assign"><bool>false</bool></edit>
  </match>
</fontconfig>
EOF
        } > "$conf.tmp"
        mv "$conf.tmp" "$conf"
    fi
    export FONTCONFIG_FILE="$conf"
    export _TAKIGO_FONTS_PINNED=1
}

# LLM_TOOL — selects the LLM used by fix_demo.sh / fix_all.sh for auto-fixes.
# It points to a predefined id (see LLM_TOOLS below); each id maps to a full
# launch command. The prompt is piped on stdin; the tool must write its
# response to stdout. Example:
#   LLM_TOOL=opencode-deepseek-v4-pro bash scripts/fix_demo.sh button
LLM_TOOL="${LLM_TOOL:-claude}"

# Predefined LLM launch configs, keyed by id. Add new ids here to make them
# selectable via LLM_TOOL.
declare -A LLM_TOOLS=(
    [claude]='claude -p /dev/stdin --allowedTools "Read,Edit,Bash" --permission-mode bypassPermissions --output-format text'
    [opencode-deepseek-v4-pro]='opencode run --model deepseek/deepseek-v4-pro --auto'
    [opencode-minimax-m3]='opencode run --model minimax-coding-plan/MiniMax-M3 --auto'
)

# Resolve LLM_TOOL to LLM_COMMAND (the launch command) and LLM_NAME (the binary
# name, used for the guard and progress messages).
if [[ -z "${LLM_TOOLS[$LLM_TOOL]:-}" ]]; then
    echo "ERROR: unknown LLM_TOOL '$LLM_TOOL'." >&2
    echo "Known tools: $(printf '%s ' "${!LLM_TOOLS[@]}")" >&2
    exit 1
fi
LLM_COMMAND="${LLM_TOOLS[$LLM_TOOL]}"
LLM_NAME="${LLM_COMMAND%% *}"

# llm_guard — refuse to run the claude CLI from inside a Claude Code session
# (claude-in-claude is problematic). Other tools (e.g. opencode) are allowed.
llm_guard() {
    if [[ "$LLM_NAME" == claude && -n "${CLAUDECODE:-}" ]]; then
        echo "ERROR: $0 refuses to run the 'claude' CLI from inside a Claude Code session." >&2
        echo "Set LLM_TOOL to another tool (e.g. \"opencode-deepseek-v4-pro\") or run from a regular terminal." >&2
        exit 1
    fi
}

# demotitle_bin
# Builds internal/cmd/demotitle once into tmp/bin and echoes the path. The cached binary
# is rebuilt only when its source is newer, so the per-demo `go run` cost is
# paid once per source change instead of once per invocation.
demotitle_bin() {
    local bin="$BIN_DIR/demotitle"
    local src="$PROJECT_DIR/internal/cmd/demotitle/main.go"
    if [[ ! -x "$bin" || "$src" -nt "$bin" ]]; then
        (cd "$PROJECT_DIR" && go build -o "$bin" ./internal/cmd/demotitle) || return 1
    fi
    echo "$bin"
}

# demodiff_bin
# Builds internal/cmd/demodiff once into tmp/bin (rebuilt when its sources or
# internal/treedump change) and echoes the path.
demodiff_bin() {
    local bin="$BIN_DIR/demodiff"
    local newer
    if [[ -x "$bin" ]]; then
        newer=$(find "$PROJECT_DIR/internal/cmd/demodiff" "$PROJECT_DIR/internal/treedump" -name '*.go' -newer "$bin" | head -1)
    fi
    if [[ ! -x "$bin" || -n "$newer" ]]; then
        (cd "$PROJECT_DIR" && go build -o "$bin" ./internal/cmd/demodiff) || return 1
    fi
    echo "$bin"
}

# find_windows_exact TITLE
# Prints window IDs whose WM_NAME exactly equals TITLE (fixed string, not
# regex). Window titles may contain regex metacharacters (e.g. parentheses),
# so xdotool search --name (which treats the pattern as a regex) cannot be
# used directly for exact matching.
find_windows_exact() {
    local title="$1"
    local id
    xdotool search --name "" 2>/dev/null | while read -r id; do
        [[ -z "$id" ]] && continue
        [[ "$(xdotool getwindowname "$id" 2>/dev/null)" == "$title" ]] && echo "$id"
    done || true
}

# odiff_score BASE COMPARE [DIFF_OUT]
# Runs odiff with anti-aliasing ignored and prints the diff percentage
# (0-100, lower = more similar). Exit 0 on success, 1 on error.
# odiff exit codes: 0 = match, 21 = layout diff, 22 = pixel diff.
odiff_score() {
    local base="$1" cmp="$2" diff_out="${3:-}"
    local out rc
    [[ -n "$diff_out" ]] && rm -f "$diff_out"
    out=$(odiff "$base" "$cmp" ${diff_out:+"$diff_out"} --parsable-stdout --aa 2>/dev/null)
    rc=$?
    case "$rc" in
        0)  # odiff >= 4 writes no diff image for identical inputs.
            [[ -n "$diff_out" && ! -f "$diff_out" ]] && cp "$base" "$diff_out"
            echo "0"; return 0 ;;
        22) echo "${out##*;}"; return 0 ;;
        21) echo "layout-diff" >&2; return 1 ;;
        *)  echo "odiff failed (exit $rc)" >&2; return 1 ;;
    esac
}

# tcl_demo_for GO_DEMO
# Prints the Tcl counterpart name. Prints "-" and returns 1 if Go-only or
# unknown. Sources demo_map.sh on first call so the mapping is loaded.
if [[ -z "${_DEMO_MAP_LOADED:-}" ]]; then
    # shellcheck disable=SC1091
    source "$SCRIPT_DIR/demo_map.sh"
    _DEMO_MAP_LOADED=1
fi
tcl_demo_for() {
    local go_demo="$1"
    if [[ -z "${DEMO_MAP[$go_demo]:-}" || "${DEMO_MAP[$go_demo]}" == "-" ]]; then
        echo "-"
        return 1
    fi
    echo "${DEMO_MAP[$go_demo]}"
}

# extract_demo_title GO_DEMO
# Reads the Title() option from the Go source via internal/cmd/demotitle.
# Returns the empty string if the demo has no Title() call or the file
# can't be parsed. Use go/parser + go/ast (not grep) so escaped strings
# and comments are handled correctly.
extract_demo_title() {
    local demo="$1"
    local src="$PROJECT_DIR/demos/$demo/main.go"
    local bin
    bin=$(demotitle_bin) 2>/dev/null || true
    [[ -z "$bin" ]] && return 0
    "$bin" "$src" 2>/dev/null || true
}

# extract_demo_geometry GO_DEMO
# Reads the Geometry() option from the Go source so the Tcl wrapper can be
# positioned identically. Falls back to "+300+300" when the demo has no
# Geometry() call (matching the wrapper's default).
extract_demo_geometry() {
    local demo="$1"
    local src="$PROJECT_DIR/demos/$demo/main.go"
    local geom
    local bin
    bin=$(demotitle_bin) 2>/dev/null || true
    [[ -z "$bin" ]] && { echo "+300+300"; return 0; }
    geom=$("$bin" -geometry "$src") 2>/dev/null || geom=""
    [[ -z "$geom" ]] && geom="+300+300"
    echo "$geom"
}

# set_skip_if_exists FLAG (0 or 1)
# 0 = always retake screenshots; 1 = reuse existing files when present.
set_skip_if_exists() {
    if [[ "$1" == "0" ]]; then
        export SKIP_IF_EXISTS=0
    else
        export SKIP_IF_EXISTS=1
    fi
}

# run_compare GO_DEMO [TCL_DEMO]
# Runs demo_compare.sh for the given Go demo (and its Tcl counterpart, or
# the second argument if it differs). Captures the score and prints it on
# stdout. Stderr goes to a per-demo log file in tmp/logs/ so failures
# are diagnosable instead of silently discarded.
#
# Exit codes:
#   0  — comparison succeeded; the score is on stdout.
#   1  — demo_compare.sh failed (see the per-demo log).
run_compare() {
    local go_demo="$1"
    local tcl_demo="${2:-}"
    if [[ -z "$tcl_demo" ]]; then
        tcl_demo=$(tcl_demo_for "$go_demo") || {
            echo "no Tcl counterpart for $go_demo" >&2
            return 1
        }
    fi
    local err_log="$LOGS_DIR/${go_demo}.compare.err"
    local out
    out=$(bash "$SCRIPT_DIR/demo_compare.sh" "$go_demo" "$tcl_demo" 2>"$err_log") || {
        echo "compare failed for $go_demo (see $err_log)" >&2
        return 1
    }
    rm -f "$err_log"
    # Score line format: "Diff score: <number>  (odiff diff % ...)"
    echo "$out" | awk '/^Diff score:/ {print $3}'
}

# update_baseline SCORES_TSV
# Merges a demo_batch.sh scores.tsv into the committed demos/parity.tsv:
# processed demos get new numbers and status, others and all notes are kept.
update_baseline() {
    local scores="$1"
    local baseline="$PROJECT_DIR/demos/parity.tsv"
    local tmp_base summary
    tmp_base="$baseline.tmp"
    awk -F'\t' -v OFS='\t' '
        FNR == 1 && FILENAME == ARGV[1] { next }
        FILENAME == ARGV[1] {
            st = ($3 == "9999") ? "failed" : (($3 == "0" && $4 == "0") ? "exact" : "close")
            new[$1] = $1 OFS $2 OFS $3 OFS $4 OFS $5 OFS $6 OFS st
            next
        }
        /^#/ || NF < 7 { next }
        { old[$1] = $0; note[$1] = $NF }
        END {
            for (d in old) if (!(d in new)) rows[d] = old[d]
            for (d in new) rows[d] = new[d] OFS note[d]
            for (d in rows) print rows[d]
        }' "$scores" "$( [[ -f "$baseline" ]] && echo "$baseline" || echo /dev/null )" \
        | LC_ALL=C sort -t$'\t' -k1,1 > "$tmp_base.rows"
    summary=$(awk -F'\t' '{c[$7]++} END {printf "exact %d / close %d / failed %d", c["exact"], c["close"], c["failed"]}' "$tmp_base.rows")
    {
        echo "# Demo parity baseline -- written by: bash scripts/demo_batch.sh --retake --update-baseline (or demo_gate.sh --accept)"
        echo "# Headless Xvfb 96dpi, pinned DejaVu fonts, frozen timers. pixel_pct = odiff % (lower is better);"
        echo "# tree_diffs = internal/cmd/demodiff structural differences; exact = both 0."
        echo "# summary: $summary"
        printf '# demo\ttcl\tpixel_pct\ttree_diffs\tgo_size\ttcl_size\tstatus\tnote\n'
        cat "$tmp_base.rows"
    } > "$tmp_base"
    rm -f "$tmp_base.rows"
    mv "$tmp_base" "$baseline"
    echo "Baseline updated: $baseline ($summary)"
}
