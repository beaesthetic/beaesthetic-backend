package it.beaesthetic.gateway.configuration

import it.beaesthetic.gateway.auth.FirestoreUserRoles
import org.springframework.aot.hint.ExecutableMode
import org.springframework.aot.hint.MemberCategory
import org.springframework.aot.hint.RuntimeHints
import org.springframework.aot.hint.RuntimeHintsRegistrar
import org.springframework.util.ReflectionUtils
import java.util.logging.Logger

class RuntimeHints : RuntimeHintsRegistrar {

    private val logger = Logger.getLogger("RuntimeHints")

    override fun registerHints(hints: RuntimeHints, classLoader: ClassLoader?) {
        getRuntimeHintClasses()
            .forEach {
                logger.info("Register type ${it.name}")
                hints.reflection().registerType(it, *MemberCategory.values())
                ReflectionUtils.getDeclaredMethods(it)
                    .forEach { method -> hints.reflection().registerMethod(method, ExecutableMode.INVOKE) }
            }

        getRuntimeHintClasses()
            .mapNotNull { it.constructors.firstOrNull() }
            .forEach {
                logger.info("Register constructor for ${it.name}")
                hints.reflection().registerConstructor(it, ExecutableMode.INVOKE)
            }
    }

    private fun getRuntimeHintClasses(): Set<Class<*>> {
        return setOf(
            FirestoreUserRoles.User::class.java,
        )
    }
}