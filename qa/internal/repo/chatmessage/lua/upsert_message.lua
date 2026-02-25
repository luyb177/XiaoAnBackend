-- KEYS[1] = list key
-- ARGV[1] = message json
-- ARGV[2] = max list size
-- ARGV[3] = ttl seconds

local key = KEYS[1]
local newMsgJson = ARGV[1]
local maxSize = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

-- 解析新消息，拿到 message_id
local ok, newMsg = pcall(cjson.decode, newMsgJson)
if not ok or not newMsg or not newMsg["message_id"] then
    return redis.error_reply("invalid message json")
end

local targetId = newMsg["message_id"]

-- ===== 扫描并删除旧消息 =====
local list = redis.call("LRANGE", key, 0, -1)

for i = 1, #list do
    local item = list[i]
    local ok2, obj = pcall(cjson.decode, item)
    if ok2 and obj and obj["message_id"] == targetId then
        -- 标记删除（Redis Lua 下标从 1 开始，但 LSET 用 0-based）
        redis.call("LSET", key, i - 1, "__DELETED__")
    end
end

-- 真正删除
redis.call("LREM", key, 0, "__DELETED__")

-- ===== 插入新消息 =====
redis.call("LPUSH", key, newMsgJson)

-- ===== 裁剪长度 =====
redis.call("LTRIM", key, 0, maxSize - 1)

-- ===== 设置 TTL =====
if ttl and ttl > 0 then
    redis.call("EXPIRE", key, ttl)
end

return 1