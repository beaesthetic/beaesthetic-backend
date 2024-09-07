package it.beaesthetic.appointment.agenda.infra.mongo

import it.beaesthetic.appointment.agenda.domain.event.CancelReason
import it.beaesthetic.appointment.agenda.domain.event.CustomerCancel
import it.beaesthetic.appointment.agenda.domain.event.NoReason
import org.bson.BsonReader
import org.bson.BsonWriter
import org.bson.codecs.Codec
import org.bson.codecs.DecoderContext
import org.bson.codecs.EncoderContext
import org.bson.codecs.configuration.CodecProvider
import org.bson.codecs.configuration.CodecRegistry

class CancelReasonCodec : Codec<CancelReason> {
    override fun encode(
        writer: BsonWriter?,
        cancelReason: CancelReason?,
        encoder: EncoderContext?
    ) {
        if (cancelReason != null) {
            writer?.writeString(
                when (cancelReason) {
                    CustomerCancel -> "CUSTOMER_CANCEL"
                    NoReason -> "NO_REASON"
                }
            )
        }
    }

    override fun getEncoderClass(): Class<CancelReason> = CancelReason::class.java

    override fun decode(reader: BsonReader?, decoderContext: DecoderContext?): CancelReason? {
        return when (reader?.readString()) {
            "CUSTOMER_CANCEL" -> return CustomerCancel
            "NO_REASON" -> return NoReason
            else -> null
        }
    }
}

class CancelReasonCodecProvider : CodecProvider {
    override fun <T : Any?> get(clazz: Class<T>?, registry: CodecRegistry?): Codec<T>? {
        if (clazz == CancelReason::class.java) {
            return CancelReasonCodec() as Codec<T>
        }

        return null
    }
}
