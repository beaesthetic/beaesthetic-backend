package it.beaesthetic.appointment.service.ports

import io.quarkus.cache.CacheKey
import io.quarkus.cache.CacheResult
import io.quarkus.runtime.annotations.RegisterForReflection
import io.smallrye.mutiny.Uni
import it.beaesthetic.appointment.generated.api.ServicesApi
import it.beaesthetic.appointment.generated.api.model.CreateServiceRequestDto
import it.beaesthetic.appointment.generated.api.model.ServiceDto
import it.beaesthetic.appointment.generated.api.model.UpdateServiceRequestDto
import it.beaesthetic.appointment.service.common.uniWithScope
import it.beaesthetic.appointment.service.domain.AppointmentService
import it.beaesthetic.appointment.service.domain.AppointmentServiceRepository
import it.beaesthetic.appointment.service.domain.Color
import java.util.*
import kotlin.NoSuchElementException

@RegisterForReflection(targets = [ServiceDto::class], registerFullHierarchy = true)
class ApplicationServiceController(
    private val applicationServiceRepository: AppointmentServiceRepository
) : ServicesApi {

    override fun createService(createServiceRequestDto: CreateServiceRequestDto): Uni<ServiceDto> =
        uniWithScope {
            val service =
                AppointmentService(
                    id = UUID.randomUUID().toString(),
                    name = createServiceRequestDto.name,
                    price = createServiceRequestDto.price.toDouble(),
                    tags = createServiceRequestDto.tags?.toSet() ?: emptySet(),
                    color = createServiceRequestDto.color?.let { Color(it) }
                )

            applicationServiceRepository
                .save(service)
                .map {
                    ServiceDto(
                        id = it.id,
                        name = it.name,
                        price = it.price.toBigDecimal(),
                        tags = it.tags.toList(),
                        color = it.color?.hex
                    )
                }
                .getOrThrow()
        }

    @CacheResult(cacheName = "all-services")
    override fun getAllServices(): Uni<List<ServiceDto>> = uniWithScope {
        applicationServiceRepository.findAll().map {
            ServiceDto(
                id = it.id,
                name = it.name,
                price = it.price.toBigDecimal(),
                tags = it.tags.toList(),
                color = it.color?.hex
            )
        }
    }

    @CacheResult(cacheName = "services-search")
    override fun searchService(
        @CacheKey text: String?,
        @CacheKey limit: Int
    ): Uni<List<ServiceDto>> = uniWithScope {
        applicationServiceRepository.searchByName(text ?: "", limit).map {
            ServiceDto(
                id = it.id,
                name = it.name,
                price = it.price.toBigDecimal(),
                tags = it.tags.toList(),
                color = it.color?.hex
            )
        }
    }

    override fun updateService(
        serviceId: String,
        updateServiceRequestDto: UpdateServiceRequestDto
    ): Uni<ServiceDto> = uniWithScope {
        val appointmentService =
            applicationServiceRepository.findById(serviceId)?.let {
                it.copy(
                    price = updateServiceRequestDto.price?.toDouble() ?: it.price,
                    tags = updateServiceRequestDto.tags?.toSet() ?: it.tags,
                    color = updateServiceRequestDto.color?.let { c -> Color(c) } ?: it.color
                )
            }
                ?: throw NoSuchElementException("Service with id $serviceId not found")
        applicationServiceRepository
            .save(appointmentService)
            .map {
                ServiceDto(
                    id = it.id,
                    name = it.name,
                    price = it.price.toBigDecimal(),
                    tags = it.tags.toList(),
                    color = it.color?.hex
                )
            }
            .getOrThrow()
    }
}
