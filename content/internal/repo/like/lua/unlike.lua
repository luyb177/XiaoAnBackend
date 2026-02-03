-- KEYS[1] = user like key     like:user:{userId}
-- KEYS[2] = target like key   like:target:{type}:{id}
-- ARGV[1] = member            type:id
-- ARGV[2] = userId

-- 如果本来就没点赞
if redis.call("SISMEMBER", KEYS[1], ARGV[1]) == 0 then
    return 0
end

-- 删除 user 维度
redis.call("SREM", KEYS[1], ARGV[1])

-- 删除 target 维度
redis.call("SREM", KEYS[2], ARGV[2])

return 1
