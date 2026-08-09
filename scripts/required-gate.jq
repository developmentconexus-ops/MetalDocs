# The key set is exactly the four upstream jobs, and every result is
# "success". No "skipped is green" allowance: whenever this gate can pass,
# `verify` was "success" — and test-integration's `needs: [verify]` with no
# `if:` means it only runs (and can reach "success") when verify succeeded.
# So the old conditional acceptance of "skipped" is dead by construction: a
# skip anywhere now fails the gate like a failure would, which is exactly
# what closes the "green over zero verification" hole (R3, ci.yml
# restructure) — a docs-only PR still runs every job, `--changed` just
# selects zero checks inside verify/test-integration and both still report
# "success", so this gate is never fooled by an honest zero-check run either.
(keys | sort) == (["lint-go", "security", "test-integration", "verify"] | sort)
and all(to_entries[]; .value.result == "success")
