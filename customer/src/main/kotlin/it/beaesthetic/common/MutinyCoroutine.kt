package it.beaesthetic.common

import io.smallrye.mutiny.coroutines.uni
import io.vertx.core.Vertx
import io.vertx.kotlin.coroutines.dispatcher
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi

private val coroutineScope: CoroutineScope by lazy {
    CoroutineScope(Vertx.currentContext().dispatcher())
}

@OptIn(ExperimentalCoroutinesApi::class)
fun <T> uniWithScope(block: suspend () -> T) = uni(coroutineScope, block)
