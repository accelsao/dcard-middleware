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

// EXAT need >= 6.2, but miniredis is 6.0
// TODO: Test with EXPIRE when API is avalible
const rateLimitScriptwithEX = `
-- KEYS[1]: IP address
local IP = KEYS[1]
local timeNow = tonumber(ARGV[1])
local ipLimit = tonumber(ARGV[2])
local duration = tonumber(ARGV[3])
local result = {}
if redis.call('EXISTS', IP) > 0 then
    local num = redis.call('INCRBY', IP, -1)
    if num < 0 then
        result[1] = -1
    else
        result[1] = num
    end
else
    redis.call('SET', IP, ipLimit - 1, 'EXAT', timeNow + duration)
    result[1] = ipLimit -1
end
result[2] = redis.call('TTL', IP)
return result
`

