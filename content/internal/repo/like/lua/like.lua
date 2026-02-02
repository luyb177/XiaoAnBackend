-- KEYS[1] = user like key     like:user:{userId}
-- KEYS[2] = target like key   like:target:{type}:{id}
-- ARGV[1] = member            type:id
-- ARGV[2] = userId

-- 已点赞
if redis.call("SISMEMBER", KEYS[1], ARGV[1]) == 1 then
    return 0
end

-- 写入 user 维度
redis.call("SADD", KEYS[1], ARGV[1])

-- 写入 target 维度
redis.call("SADD", KEYS[2], ARGV[2])

return 1
