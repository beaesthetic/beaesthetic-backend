package it.beaesthetic.customer.infra

import com.mongodb.client.model.Collation
import com.mongodb.client.model.CollationStrength
import io.quarkus.mongodb.FindOptions
import io.quarkus.mongodb.panache.kotlin.reactive.ReactivePanacheMongoRepository
import io.quarkus.panache.common.Sort
import io.smallrye.mutiny.coroutines.awaitSuspending
import it.beaesthetic.common.DomainEventRegistryDelegate
import it.beaesthetic.common.SearchGram
import it.beaesthetic.customer.application.CustomerReadRepository
import it.beaesthetic.customer.domain.*
import it.beaesthetic.customer.infra.CustomerRepositoryImpl.EntityMapper.toDomain
import jakarta.enterprise.context.ApplicationScoped
import java.time.Instant
import org.bson.Document
import org.bson.conversions.Bson

@ApplicationScoped class PanacheCustomerRepository : ReactivePanacheMongoRepository<CustomerEntity>

@ApplicationScoped
class CustomerRepositoryImpl(
    private val panacheCustomerRepository: PanacheCustomerRepository,
    private val outboxRepository: OutboxRepository<CustomerEvent>
) : CustomerRepository, CustomerReadRepository {

    override suspend fun findById(id: CustomerId): Customer? {
        return panacheCustomerRepository
            .find("_id", id.value)
            .firstResult()
            .map { entity -> entity?.let { toDomain(entity) } }
            .awaitSuspending()
    }

    override suspend fun findAll(): List<Customer> {
        return panacheCustomerRepository.findAll().list().awaitSuspending().map { toDomain(it) }
    }

    override suspend fun save(customer: Customer): Customer {
        val searchTerms =
            listOf(customer.name, customer.surname)
                .filter { it.trim().isNotEmpty() }
                .joinToString(" ")
        return panacheCustomerRepository
            .persistOrUpdate(
                CustomerEntity(
                    id = customer.id.value,
                    name = customer.name,
                    surname = customer.surname,
                    email = customer.contacts.email?.value,
                    phone = customer.contacts.phone?.fullNumber,
                    note = customer.note,
                    updatedAt = Instant.now(),
                    searchGrams =
                        SearchGram.Default.nGrams(text = searchTerms, minLength = 3)
                            .joinToString(" ")
                )
            )
            .awaitSuspending()
            .let { customer }
            .also { outboxRepository.save(customer) }
    }

    override suspend fun findNextPage(
        pageToken: String?,
        limit: Int?,
        sortBy: List<String>,
        sortDirection: Sort.Direction
    ): CustomerReadRepository.CustomerPage {
        val pageSize = limit ?: 50
        var pageBuilder = MongoPageBuilder(sortBy + "_id", sortDirection)
        val filter =
            when {
                pageToken != null -> pageBuilder.generateCursorFilter(pageToken)
                else -> Document()
            }
        var sort = pageBuilder.generateSortDocument()
        var collation =
            Collation.builder()
                .locale("en")
                .caseLevel(false)
                .collationStrength(CollationStrength.SECONDARY)
                .build()
        val customers =
            panacheCustomerRepository
                .mongoCollection()
                .find(FindOptions().filter(filter).sort(sort).limit(pageSize).collation(collation))
                .map { toDomain(it) }
                .collect()
                .asList()
                .awaitSuspending()

        return CustomerReadRepository.CustomerPage(
            customers,
            customers.size,
            nextToken =
                if (customers.size < pageSize) null
                else pageBuilder.getNextPageToken(customers.last().id.value, customers.last())
        )
    }

    override suspend fun findByKeyword(keyword: String, maxResults: Int): List<Customer> {
        val fieldsToProject = listOf("note", "phone", "surname", "name", "email", "updatedAt")
        val filter: Bson = Document("\$text", Document("\$search", keyword))
        val project: Bson =
            Document("score", Document("\$meta", "textScore")).apply {
                fieldsToProject.forEach { field -> append(field, 1L) }
            }
        val sort: Bson = Document("score", Document("\$meta", "textScore"))

        return panacheCustomerRepository
            .mongoCollection()
            .find(FindOptions().filter(filter).projection(project).sort(sort).limit(maxResults))
            .map { toDomain(it) }
            .collect()
            .asList()
            .awaitSuspending()
    }

    private object EntityMapper {
        fun toDomain(entity: CustomerEntity) =
            Customer(
                id = CustomerId(entity.id),
                name = entity.name,
                surname = entity.surname ?: "",
                contacts =
                    Contacts(
                        email = entity.email?.let { Email(it) },
                        phone =
                            entity.phone?.let { if (it.trim().isBlank()) null else Phone.of(it) },
                    ),
                note = entity.note,
                domainEventRegistry = DomainEventRegistryDelegate()
            )
    }
}
