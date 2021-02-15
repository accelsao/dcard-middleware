package middleware

// rateLimitScript is script for ratelimiter
const rateLimitScript = `
-- KEYS[1]: IP address
-- duration (second)
local IP = KEYS[1]
local timeNow = tonumber(ARGV[1])
local ipLimit = tonumber(ARGV[2])
local duration = tonumber(ARGV[3])
local userInfo = redis.call('HGETALL', IP)
local reset = tonumber(userInfo[4])
local result = {}

if #userInfo == 0 or timeNow > reset then
    reset = timeNow + duration
    redis.call('HMSET', IP, "count", 1, "reset", reset)
    result[1] = ipLimit - 1
    result[2] = reset
else
    local count = tonumber(userInfo[2])
    if count < ipLimit then
        local new_count = redis.call('HINCRBY', IP, "count", 1)
        result[1] = ipLimit - new_count
        result[2] = reset
    else
        result[1] = -1;
        result[2] = reset
    end
end
return result
`
