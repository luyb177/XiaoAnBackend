-- KEYS[1]: retry zset
-- KEYS[2]: pending list
-- ARGV[1]: now (unix timestamp)
-- ARGV[2]: limit (optional)

local limit = ARGV[2] or 100

local tasks = redis.call(
	"ZRANGEBYSCORE",
	KEYS[1],
	"-inf",
	ARGV[1],
	"LIMIT",
	0,
	limit
)

for i, task in ipairs(tasks) do
	redis.call("ZREM", KEYS[1], task)
	redis.call("LPUSH", KEYS[2], task)
end

return #tasks
