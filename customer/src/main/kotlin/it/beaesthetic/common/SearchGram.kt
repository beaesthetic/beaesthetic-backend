package it.beaesthetic.common

class SearchGram {
    companion object {
        val Default: SearchGram by lazy { SearchGram() }
    }

    fun nGrams(text: String, minLength: Int = 3): Set<String> {
        val trimmedText = text.trim()
        return (minLength until trimmedText.length)
            .step(1)
            .flatMap { nGrams(it, true)(trimmedText) }
            .toSet()
    }

    private fun nGrams(n: Int, prefix: Boolean = false): ((String) -> List<String>) {
        assert(n < 1) { "N is not valid argument" }
        return { value ->
            val index = value.length - n + 1
            (0 until index).map { value.slice(IntRange(if (prefix) 0 else it, it + n - 1)) }
        }
    }
}
