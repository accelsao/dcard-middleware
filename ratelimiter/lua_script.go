package middleware

// rateLimitScript is script for ratelimiter
const rateLimitScript = `
-- KEYS[1]: IP address
-- ARGV[1]: current time
-- ARGV[2]: ipLimit
-- ARGV[3]: duration

-- HASH: KEYS[1]
-- remains: availble times
-- reset: time for reset

-- RESULT:
-- [1] remains: remains requeset times for particular IP
-- [2] reset: reset time for particular IP
-- [3] duration: for debugging

local IP = KEYS[1]
local timeNow = tonumber(ARGV[1])
local ipLimit = tonumber(ARGV[2])
local duration = tonumber(ARGV[3])
local userInfo = redis.call('HGETALL', IP)
local reset = tonumber(userInfo[4])
local result = {}

-- print(IP, redis.call('PTTL', IP))

if #userInfo == 0 then
    result[1] = ipLimit - 1
    result[2] = timeNow + duration
    result[3] = duration

    redis.call('HSET', IP, 'remains', result[1], 'reset', result[2])
    redis.call('PEXPIRE', IP, duration)
else
    local newRemains = redis.call('HINCRBY', IP, 'remains', -1)

    if newRemains < 0 then
        newRemains = -1
    end

    result[1] = newRemains
    result[2] = reset
    result[3] = duration
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
