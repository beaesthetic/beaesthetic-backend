package it.beaesthetic.common

import io.quarkus.cache.CacheManager
import io.smallrye.mutiny.coroutines.awaitSuspending
import io.smallrye.mutiny.coroutines.uni
import io.vertx.core.Vertx
import io.vertx.kotlin.coroutines.dispatcher
import jakarta.enterprise.context.ApplicationScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.SupervisorJob

@ApplicationScoped
class SuspendableCache(private val vertx: Vertx, private val cacheManager: CacheManager) {

    private val coroutineScope: CoroutineScope by lazy {
        CoroutineScope(vertx.dispatcher() + SupervisorJob())
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    suspend fun <V> getOrLoad(cacheName: String, vararg keys: Any?, loader: suspend () -> V): V {
        val cache = cacheManager.getCache(cacheName).orElseThrow()
        val cacheKey = keys.filterNotNull().joinToString(":")
        return cache.getAsync(cacheKey) { uni(coroutineScope, loader) }.awaitSuspending()
    }

    suspend fun invalidateAll(cacheName: String) {
        val cache = cacheManager.getCache(cacheName).orElseThrow()
        cache.invalidateAll().awaitSuspending()
    }

    suspend fun invalidateKey(cacheName: String, vararg keys: Any?) {
        val cacheKey = keys.filterNotNull().joinToString(":")
        val cache = cacheManager.getCache(cacheName).orElseThrow()
        cache.invalidate(cacheKey).awaitSuspending()
    }
}
