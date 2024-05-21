package it.beaesthetic.gateway.auth

import arrow.core.Either
import com.google.api.core.ApiFuture
import com.google.api.core.ApiFutureToListenableFuture
import com.google.firebase.auth.FirebaseAuth
import com.google.firebase.auth.FirebaseAuthException
import com.google.firebase.auth.FirebaseToken
import kotlinx.coroutines.guava.await
import kotlinx.coroutines.reactor.mono
import org.springframework.http.HttpStatus
import org.springframework.security.authentication.ReactiveAuthenticationManager
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken
import org.springframework.security.core.Authentication
import org.springframework.web.server.ResponseStatusException
import reactor.core.publisher.Mono

class FirebaseAuthenticationManager(private val firebaseAuth: FirebaseAuth) : ReactiveAuthenticationManager {

    override fun authenticate(authentication: Authentication?): Mono<Authentication> {
        val idToken = authentication?.credentials.toString()
        return mono { verifyFirebaseToken(idToken) }
            .flatMap { it.fold(
                { error -> Mono.error(ResponseStatusException(HttpStatus.UNAUTHORIZED, "Invalid token", error)) },
                { decodedToken -> Mono.just(decodedToken) }
            ) }
            .map { User(it.uid, it?.name ?: it.email, it.email) }
            .map { UsernamePasswordAuthenticationToken.authenticated(it, authentication?.credentials, emptyList()) }
    }

    private suspend fun verifyFirebaseToken(token: String): Either<Error, FirebaseToken> {
        return try {
            Either.Right(firebaseAuth.verifyIdTokenAsync(token).await())
        } catch (e: FirebaseAuthException) {
            return Either.Left(Error(e))
        }
    }

}

suspend fun <T> ApiFuture<T>.await(): T {
    return ApiFutureToListenableFuture(this).await()
}