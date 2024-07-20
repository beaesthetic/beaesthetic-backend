package it.beaesthetic.customer.domain

@JvmInline
value class Email(val value:String)

@JvmInline
value class Phone(val value:String)

data class Contacts(
    val email: Email?,
    val phone: Phone?,
)