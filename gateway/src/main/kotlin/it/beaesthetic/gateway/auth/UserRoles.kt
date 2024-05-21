package it.beaesthetic.gateway.auth

typealias Role = String

interface UserRoles {
    suspend fun findByUserEmail(email: String): Set<Role>

    suspend fun findByUserId(userId: String): Set<Role>
}