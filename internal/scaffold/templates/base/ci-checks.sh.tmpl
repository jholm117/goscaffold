#!/usr/bin/env bash
set -euo pipefail

parallel=false
for arg in "$@"; do
    case "$arg" in
        --parallel) parallel=true ;;
    esac
done

failed=0

run_lint() {
    echo "--- make lint-config"
    make lint-config
    echo "--- make lint"
    make lint
}

run_tidy_check() {
    echo "--- go mod tidy (checking for drift)"
    cp go.mod go.mod.bak
    cp go.sum go.sum.bak
    go mod tidy
    if ! diff -q go.mod go.mod.bak >/dev/null 2>&1 || ! diff -q go.sum go.sum.bak >/dev/null 2>&1; then
        echo "ERROR: go.mod or go.sum changed after 'go mod tidy'. Commit the changes."
        mv go.mod.bak go.mod
        mv go.sum.bak go.sum
        return 1
    fi
    rm -f go.mod.bak go.sum.bak
}

run_test() {
    echo "--- make test"
    make test
}

run_govulncheck() {
    echo "--- make govulncheck"
    make govulncheck
}

run_tidy_check || { echo "FAIL: go mod tidy"; exit 1; }

if [ "$parallel" = true ]; then
    echo "==> installing tools before parallel run"
    make tools
    echo "==> running checks in parallel"
    run_lint &
    lint_pid=$!
    run_test &
    test_pid=$!
    run_govulncheck &
    vuln_pid=$!

    if ! wait "$lint_pid"; then echo "FAIL: lint"; failed=1; fi
    if ! wait "$test_pid"; then echo "FAIL: test"; failed=1; fi
    if ! wait "$vuln_pid"; then echo "FAIL: govulncheck"; failed=1; fi
else
    echo "==> running checks sequentially"
    run_lint || { echo "FAIL: lint"; failed=1; }
    run_test || { echo "FAIL: test"; failed=1; }
    run_govulncheck || { echo "FAIL: govulncheck"; failed=1; }
fi

if [ "$failed" -ne 0 ]; then exit 1; fi
echo "==> all checks passed"
