package it.beaesthetic.customer.domain

import java.util.regex.Pattern

@JvmInline value class Email(val value: String)

data class Contacts(
    val email: Email?,
    val phone: Phone?,
)

data class Phone(val prefix: String, val number: String) {
    val fullNumber: String
        get() = "${prefix}${number}"

    companion object {
        private val PREFIX_REGEX = Pattern.compile("^\\+?[0-9]{2}")

        fun of(fullPhone: String): Phone {
            val prefixMatch = PREFIX_REGEX.matcher(fullPhone)
            require(prefixMatch.find()) { "Invalid phone number format: $fullPhone" }
            val prefix = prefixMatch.group()
            val number = fullPhone.substring(prefix.length)
            return Phone(prefix, number)
        }
    }
}
