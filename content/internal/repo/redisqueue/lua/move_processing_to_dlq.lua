-- KEYS[1] = processing list
-- KEYS[2] = dlq list
-- ARGV[1] = payload

-- 1. remove from processing
local removed = redis.call("LREM", KEYS[1], 1, ARGV[1])
if removed == 0 then
    return {err = "task not found in processing"}
end

-- 2. push to dlq
redis.call("LPUSH", KEYS[2], ARGV[1])

return "OK"
