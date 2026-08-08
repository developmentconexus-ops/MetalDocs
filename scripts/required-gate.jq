(keys | sort) == ([
  "changes","lint-go","lint-contract","lint-frontend","governance",
  "test-go","test-frontend","test-integration","security"
] | sort)
and .changes.result == "success"
and all(to_entries[];
        if .key == "changes" then .value.result == "success"
        else (.value.result == "success" or .value.result == "skipped") end)
