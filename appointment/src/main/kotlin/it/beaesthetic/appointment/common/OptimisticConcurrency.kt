package it.beaesthetic.appointment.common

object OptimisticConcurrency {
    data class VersionedEntity<T>(
        val entity: T,
        val version: Long
    )

    fun <T> T.versioned(version: Long): VersionedEntity<T> =
        VersionedEntity(this, version)
}