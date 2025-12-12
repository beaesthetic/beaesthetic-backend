package it.beaesthetic.common

import io.quarkus.cache.Cache
import io.smallrye.mutiny.coroutines.awaitSuspending
import io.smallrye.mutiny.coroutines.uni
import io.vertx.core.Vertx
import io.vertx.kotlin.coroutines.dispatcher
import jakarta.enterprise.context.ApplicationScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.SupervisorJob

@ApplicationScoped
class SuspendableCache(private val vertx: Vertx) {

    private val coroutineScope: CoroutineScope by lazy {
        CoroutineScope(vertx.dispatcher() + SupervisorJob())
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    suspend fun <V> getOrLoad(cache: Cache, vararg keys: Any?, loader: suspend () -> V): V {
        val cacheKey = keys.filterNotNull().joinToString(":")
        return cache.getAsync(cacheKey) { uni(coroutineScope, loader) }.awaitSuspending()
    }

    suspend fun <V> loadAndInvalidate(cache: Cache, vararg keys: Any?, loader: suspend () -> V): V {
        val cacheKey = keys.filterNotNull().joinToString(":")
        return loader.invoke().also { cache.invalidate(cacheKey) }
    }

    suspend fun invalidateAll(cache: Cache) {
        cache.invalidateAll().awaitSuspending()
    }

    suspend fun invalidateKey(cache: Cache, vararg keys: Any?) {
        val cacheKey = keys.filterNotNull().joinToString(":")
        cache.invalidate(cacheKey).awaitSuspending()
    }
}
