package io.github.petretiandrea.scheduler.redis

import com.fasterxml.jackson.databind.ObjectMapper
import io.github.petretiandrea.scheduler.core.Job
import io.github.petretiandrea.scheduler.core.ScheduleId
import io.github.petretiandrea.scheduler.core.ScheduleJob
import io.github.petretiandrea.scheduler.core.ScheduleJobRepository
import java.time.Duration
import java.time.Instant
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.reactive.awaitFirstOrNull
import kotlinx.coroutines.withContext
import org.springframework.data.redis.core.ReactiveRedisTemplate
import org.springframework.data.redis.core.RedisOperations
import org.springframework.data.redis.core.RedisTemplate
import org.springframework.data.redis.core.SessionCallback

data class ScheduleJobRedisOptions(
    val sortedSetName: String,
    val taskSetName: String,
    val peekBatchSize: Int,
    val peekLeaseTTL: Duration,
)

class RedisJobRepository(
    private val reactiveRedisTemplate: ReactiveRedisTemplate<String, String>,
    private val redisTemplate: RedisTemplate<String, String>,
    private val scheduleJobRedisOptions: ScheduleJobRedisOptions,
    private val objectMapper: ObjectMapper
) : ScheduleJobRepository {

    private val sortedSetExtension = SortedSetExtension(reactiveRedisTemplate)

    override suspend fun save(job: ScheduleJob): Result<ScheduleJob> =
        withContext(Dispatchers.IO) {
            runCatching {
                    redisTemplate.execute(
                        object : SessionCallback<String> {
                            private fun a(operations: RedisOperations<String, String>): String? {
                                operations.watch(
                                    listOf(
                                        scheduleJobRedisOptions.taskSetName,
                                        scheduleJobRedisOptions.sortedSetName
                                    )
                                )
                                operations.multi()
                                operations
                                    .opsForZSet()
                                    .add(
                                        scheduleJobRedisOptions.sortedSetName,
                                        job.id.id,
                                        job.scheduleAt.toEpochMilli().toDouble()
                                    )
                                operations
                                    .opsForValue()
                                    .set(job.id.id, objectMapper.writeValueAsString(job))
                                operations.exec()
                                return null
                            }

                            override fun <K : Any?, V : Any?> execute(
                                operations: RedisOperations<K, V>
                            ): String? {
                                return a(operations as RedisOperations<String, String>)
                            }
                        }
                    )
                }
                .map { job }
        }

    override suspend fun findById(id: ScheduleId): Result<ScheduleJob> {
        TODO("Not yet implemented")
    }

    override suspend fun pollJobs(tick: Instant, delta: kotlin.time.Duration): List<Job> {
        val ttl = tick + scheduleJobRedisOptions.peekLeaseTTL
        return sortedSetExtension
            .sortedSetPeekAndLease(
                sortedSetName = scheduleJobRedisOptions.sortedSetName,
                peekTime = tick,
                peekBatchSize = scheduleJobRedisOptions.peekBatchSize,
                peekUntil = ttl
            )
            .flatMap { taskId -> reactiveRedisTemplate.opsForValue().get(taskId) }
            .map { objectMapper.readValue(it, ScheduleJob::class.java) }
            .map { task -> RedisJobImpl(task, redisTemplate, scheduleJobRedisOptions) }
            .collectList()
            .awaitFirstOrNull()
            ?: emptyList()
    }

    override suspend fun remove(scheduleId: ScheduleId): Result<Unit> =
        withContext(Dispatchers.IO) {
            runCatching {
                    redisTemplate.execute(
                        object : SessionCallback<String> {
                            private fun a(operations: RedisOperations<String, String>): String? {
                                operations.watch(
                                    listOf(
                                        scheduleJobRedisOptions.taskSetName,
                                        scheduleJobRedisOptions.sortedSetName
                                    )
                                )
                                operations.multi()
                                operations
                                    .opsForZSet()
                                    .remove(scheduleJobRedisOptions.sortedSetName, scheduleId.id)
                                operations.delete(scheduleId.id)
                                operations.exec()
                                return null
                            }
                            override fun <K : Any?, V : Any?> execute(
                                operations: RedisOperations<K, V>
                            ): String? {
                                return a(operations as RedisOperations<String, String>)
                            }
                        }
                    )
                }
                .map {}
        }

    private class RedisJobImpl(
        override val scheduleJob: ScheduleJob,
        private val redisTemplate: RedisTemplate<String, String>,
        private val scheduleJobRedisOptions: ScheduleJobRedisOptions
    ) : Job {
        override suspend fun ack() =
            withContext(Dispatchers.IO) {
                redisTemplate.execute(
                    object : SessionCallback<String> {
                        private fun a(operations: RedisOperations<String, String>): String? {
                            operations.multi()
                            operations
                                .opsForZSet()
                                .remove(scheduleJobRedisOptions.sortedSetName, scheduleJob.id.id)
                            operations.delete(scheduleJob.id.id)
                            operations.exec()
                            return null
                        }
                        override fun <K : Any?, V : Any?> execute(
                            operations: RedisOperations<K, V>
                        ): String? {
                            return a(operations as RedisOperations<String, String>)
                        }
                    }
                )
                Unit
            }

        override suspend fun nack(reason: String?) = withContext(Dispatchers.IO) { TODO() }
    }
}
