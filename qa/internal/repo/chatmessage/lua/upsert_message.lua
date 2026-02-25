-- KEYS[1] = list key
-- ARGV[1] = message json
-- ARGV[2] = max size
-- ARGV[3] = ttl seconds

local key = KEYS[1]
local msgJson = ARGV[1]
local maxSize = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

-- 从新消息中提取 message_id
local ok, newMsg = pcall(cjson.decode, msgJson)
if not ok then
    return redis.error_reply("invalid json")
end

local newID = newMsg["message_id"]
if not newID then
    return redis.error_reply("missing message_id")
end

-- =============================
-- O(n) 扫描（n<=30 完全安全）
-- =============================
local list = redis.call("LRANGE", key, 0, -1)

for i = 1, #list do
    local ok2, oldMsg = pcall(cjson.decode, list[i])
    if ok2 and oldMsg["message_id"] == newID then
        -- 删除旧消息
        redis.call("LREM", key, 1, list[i])
        break
    end
end

-- 插入到头部（最新在前）
redis.call("LPUSH", key, msgJson)

-- 裁剪长度
redis.call("LTRIM", key, 0, maxSize - 1)

-- 刷新 TTL
redis.call("EXPIRE", key, ttl)

return 1