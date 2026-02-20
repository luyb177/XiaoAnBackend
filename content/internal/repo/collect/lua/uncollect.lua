-- KEYS[1] = user collect key     collect:user:{userID}
-- KEYS[2] = target collect key   collect:target:{type}:{id}
-- ARGV[1] = member               type:id
-- ARGV[2] = userID

-- 如果本来就没收藏
if redis.call("SISMEMBER", KEYS[1], ARGV[1]) == 0 then
    return 0
end

-- 删除 user 维度
redis.call("SREM", KEYS[1], ARGV[1])

-- 删除 target 维度
redis.call("SREM", KEYS[2], ARGV[2])

return 1
