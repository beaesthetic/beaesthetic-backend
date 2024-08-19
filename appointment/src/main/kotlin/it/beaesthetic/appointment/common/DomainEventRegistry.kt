package it.beaesthetic.appointment.common

interface DomainEventRegistry<E> {
    fun addEvent(eventType: String, event: E)
    fun clearEvents()
    val events: List<Pair<String, E>>

    companion object {
        fun <E> delegate(
            domainEvents: List<Pair<String, E>> = emptyList()
        ): DomainEventRegistry<E> = DomainEventRegistryDelegate(domainEvents)

        private class DomainEventRegistryDelegate<E>(
            private var domainEvents: List<Pair<String, E>> = emptyList(),
        ) : DomainEventRegistry<E> {
            override fun addEvent(eventType: String, event: E) {
                domainEvents += eventType to event
            }

            override fun clearEvents() {
                domainEvents = emptyList()
            }

            override val events: List<Pair<String, E>>
                get() = domainEvents
        }
    }
}
