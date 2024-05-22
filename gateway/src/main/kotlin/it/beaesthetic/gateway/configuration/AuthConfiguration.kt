package it.beaesthetic.gateway.configuration

import com.google.cloud.firestore.Firestore
import com.google.firebase.auth.FirebaseAuth
import it.beaesthetic.gateway.auth.FirebaseAuthenticationManager
import it.beaesthetic.gateway.auth.FirestoreUserRoles
import it.beaesthetic.gateway.auth.UserRoles
import org.springframework.beans.factory.annotation.Value
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.ImportRuntimeHints
import org.springframework.http.HttpMethod
import org.springframework.security.authentication.ReactiveAuthenticationManager
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken
import org.springframework.security.config.annotation.method.configuration.EnableReactiveMethodSecurity
import org.springframework.security.config.annotation.web.reactive.EnableWebFluxSecurity
import org.springframework.security.config.web.server.SecurityWebFiltersOrder
import org.springframework.security.config.web.server.ServerHttpSecurity
import org.springframework.security.web.server.SecurityWebFilterChain
import org.springframework.security.web.server.authentication.AuthenticationWebFilter
import org.springframework.security.web.server.authentication.ServerAuthenticationConverter
import org.springframework.web.cors.CorsConfiguration
import org.springframework.web.cors.reactive.CorsConfigurationSource
import org.springframework.web.cors.reactive.UrlBasedCorsConfigurationSource
import reactor.core.publisher.Mono


@Configuration
@EnableWebFluxSecurity
@EnableReactiveMethodSecurity(useAuthorizationManager = true)
@ImportRuntimeHints(RuntimeHints::class)
class AuthConfiguration {

    @Bean
    fun userRoles(
        firestore: Firestore,
        @Value("\${firebase.userRoleCollection}") userRoleName: String,
    ): UserRoles {
        return FirestoreUserRoles(firestore, userRoleName)
    }

    @Bean
    fun filterChain(
        http: ServerHttpSecurity,
        firebaseAuthenticationManager: ReactiveAuthenticationManager,
        bearerTokenConverter: ServerAuthenticationConverter
    ): SecurityWebFilterChain {
        val authenticationWebFilter = AuthenticationWebFilter(firebaseAuthenticationManager)
        authenticationWebFilter.setServerAuthenticationConverter(bearerTokenConverter)
        return http
            .csrf { it.disable() }
            .cors { cors -> cors.configurationSource(corsConfigurationSource())}
            .addFilterAt(authenticationWebFilter, SecurityWebFiltersOrder.AUTHENTICATION)
            .authorizeExchange {
                it.pathMatchers(HttpMethod.OPTIONS, "/**").permitAll()
                    .pathMatchers(HttpMethod.GET, "/actuator/**").permitAll()
                    .anyExchange().authenticated()
            }
            .build()
    }

    private fun corsConfigurationSource(): CorsConfigurationSource {
        val configuration = CorsConfiguration()
        configuration.allowedOrigins = listOf("*")
        configuration.setAllowedMethods(
            listOf(
                "GET", "POST", "PUT",
                "DELETE", "OPTIONS"
            )
        )
        configuration.allowCredentials = false
        configuration.allowedHeaders = listOf("*")
        val source = UrlBasedCorsConfigurationSource()
        source.registerCorsConfiguration("/**", configuration)
        return source
    }

    @Bean
    fun firebaseAuthenticationManager(firebaseAuth: FirebaseAuth): ReactiveAuthenticationManager {
        return FirebaseAuthenticationManager(firebaseAuth)
    }


    @Bean
    fun bearerTokenConverter(): ServerAuthenticationConverter {
        return ServerAuthenticationConverter { exchange ->
            val authorizationHeader = exchange?.request?.headers?.getFirst("Authorization")
                ?.let { if (it.startsWith("Bearer ")) it.substring(7) else null }

            if (authorizationHeader != null) {
                Mono.just(
                    UsernamePasswordAuthenticationToken(authorizationHeader, authorizationHeader)
                )
            } else {
                Mono.empty()
            }
        }
    }
}