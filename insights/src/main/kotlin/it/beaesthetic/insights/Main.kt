package it.beaesthetic.insights

import io.quarkus.runtime.Quarkus
import io.quarkus.runtime.QuarkusApplication
import io.quarkus.runtime.annotations.QuarkusMain
import it.beaesthetic.insights.service.InsightService
import kotlinx.coroutines.runBlocking

@QuarkusMain
class Main {

    companion object {
        @JvmStatic
        fun main(args: Array<String>) {
            Quarkus.run(App::class.java, *args)
        }
    }

    class App(
        private val insightService: InsightService
    ) : QuarkusApplication {
        override fun run(vararg args: String?): Int = runBlocking {
            insightService.computeMostUsedTreatments()
            insightService.computeTreatmentsByCustomer()
            Quarkus.asyncExit(0)
            0
        }
    }
}
