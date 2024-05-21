package it.beaesthetic.gateway.auth

import com.google.firebase.auth.FirebaseAuth
import org.springframework.http.ResponseEntity
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("auth")
@Validated
class AuthController(
    private val userRoles: UserRoles,
    private val firebaseAuth: FirebaseAuth,
){

    companion object {
        const val ROLES_KEY = "roles"
    }

    @PostMapping("/signUpFirebase")
    suspend fun signUpWithFirebase(
        @AuthenticationPrincipal user: User
    ) : ResponseEntity<Unit> {
        val roles = userRoles.findByUserEmail(user.email)
        return if (roles.isNotEmpty()) {
            firebaseAuth.setCustomUserClaimsAsync(
                user.id,
                mapOf(
                    ROLES_KEY to roles
                )
            ).await()
            ResponseEntity.ok().build()
        } else {
            ResponseEntity.notFound().build()
        }
    }
}