package it.beaesthetic.gateway.auth

import com.google.cloud.firestore.CollectionReference
import com.google.cloud.firestore.Firestore
import com.google.gson.Gson

class FirestoreUserRoles(
    firestore: Firestore,
    collectionName: String
) : UserRoles {

    private val gson: Gson = Gson()
    private val collection: CollectionReference =
        firestore.collection(collectionName)



    override suspend fun findByUserEmail(email: String): Set<Role> {
        return collection.whereEqualTo("email", email)
            .limit(1)
            .get()
            .await()
            .documents
            .firstOrNull()
            ?.let { gson.toJson(it.data) }
            ?.let { gson.fromJson(it, User::class.java) }
            ?.roles
            ?.toSet()
            .orEmpty()
    }

    override suspend fun findByUserId(userId: String): Set<Role> {
        TODO("Not yet implemented")
    }

    data class User(
        val email: String,
        val roles: List<String>
    )

}