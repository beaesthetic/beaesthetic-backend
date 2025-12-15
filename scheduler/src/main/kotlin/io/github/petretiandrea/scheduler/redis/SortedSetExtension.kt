package io.github.petretiandrea.scheduler.redis

import java.time.Instant
import org.springframework.data.redis.core.ReactiveRedisTemplate
import org.springframework.data.redis.core.script.RedisScript
import reactor.core.publisher.Flux

class SortedSetExtension(private val reactiveRedisTemplate: ReactiveRedisTemplate<String, String>) {
    fun sortedSetPeekAndLease(
        sortedSetName: String,
        peekTime: Instant,
        peekBatchSize: Int,
        peekUntil: Instant,
    ): Flux<String> {
        val script by lazy {
            val scriptSource =
                """
                local items = redis.call('ZRANGEBYSCORE', KEYS[1], ARGV[1], ARGV[2], 'LIMIT', 0, ARGV[3]);
                if #items > 0 then
                    for i, v in ipairs(items) do
                        redis.call('ZADD', KEYS[1], 'XX', ARGV[4] or 0.0, v)
                    end;
                end;
                return items;
                """
                    .trimIndent()
            RedisScript.of(scriptSource, List::class.java)
        }
        val args =
            listOf(
                Long.MIN_VALUE.toString(),
                peekTime.toEpochMilli().toString(),
                peekBatchSize.toString(),
                peekUntil.toEpochMilli().toString(),
            )
        return reactiveRedisTemplate.execute(script, listOf(sortedSetName), args).flatMap {
            Flux.fromIterable(it as List<String>)
        }
    }
}
