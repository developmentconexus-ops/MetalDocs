# The key set is exactly the four upstream jobs, and every result must be
# literally "success". No "skipped is green" allowance, and — since the
# integration suite stopped staging behind `verify` — no reliance on one
# either: this gate does not care WHY a job is not "success", so a skip, a
# cancellation and a failure are all equally fatal. That is strictly stronger
# than the argument this comment used to make, which reasoned from
# test-integration's `needs: [verify]` edge and would have needed rewriting
# every time the job graph moved.
#
# It is what closes the "green over zero verification" hole (R3, ci.yml
# restructure): a docs-only PR still runs every job, `--changed` just selects
# zero checks inside verify/test-integration and both still report "success",
# so the gate is not fooled by an honest zero-check run either.
#
# test-integration is a sharded matrix job. `needs.test-integration.result`
# is the AGGREGATE of its legs — any leg that fails or is cancelled makes the
# aggregate non-"success" — so the four shards need no representation here.
(keys | sort) == (["lint-go", "security", "test-integration", "verify"] | sort)
and all(to_entries[]; .value.result == "success")
