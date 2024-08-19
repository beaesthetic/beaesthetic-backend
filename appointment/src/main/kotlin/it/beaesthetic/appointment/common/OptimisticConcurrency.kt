package it.beaesthetic.appointment.common

data class OptimisticLockException(val version: Long) :
    Exception("Version mismatch, expected version: $version")

object OptimisticConcurrency {
    data class VersionedEntity<T>(val entity: T, val version: Long)

    fun <T> T.versioned(version: Long): VersionedEntity<T> = VersionedEntity(this, version)
}
