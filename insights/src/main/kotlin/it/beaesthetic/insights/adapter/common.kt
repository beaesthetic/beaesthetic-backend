package it.beaesthetic.insights.adapter

import io.vertx.core.Vertx
import io.vertx.kotlin.coroutines.dispatcher
import jakarta.inject.Singleton
import kotlinx.coroutines.CoroutineScope

@Singleton
val vertxCoroutineScope = CoroutineScope(Vertx.currentContext().dispatcher())