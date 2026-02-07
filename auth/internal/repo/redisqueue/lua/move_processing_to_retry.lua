-- KEYS[1] = processing list
-- KEYS[2] = retry zset
-- ARGV[1] = old payload (for LREM)
-- ARGV[2] = new payload (updated retry count)
-- ARGV[3] = score (unix timestamp)

-- 1. remove from processing
local removed = redis.call("LREM", KEYS[1], 1, ARGV[1])
if removed == 0 then
    return {err = "task not found in processing"}
end

-- 2. add to retry zset
redis.call("ZADD", KEYS[2], ARGV[3], ARGV[2])

return "OK"
