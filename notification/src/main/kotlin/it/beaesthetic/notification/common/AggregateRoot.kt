package it.beaesthetic.notification.common

interface DomainEventRegistry<E> {
    fun addEvent(event: E)
    fun clearEvents()
    val events: List<E>
}

class DomainEventRegistryDelegate<E>(
    private var domainEvents: List<E> = emptyList(),
) : DomainEventRegistry<E> {
    override fun addEvent(event: E) {
        domainEvents += event
    }

    override fun clearEvents() {
        domainEvents = emptyList()
    }

    override val events: List<E>
        get() = domainEvents
}
