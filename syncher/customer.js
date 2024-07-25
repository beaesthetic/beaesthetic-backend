const {MONGO_CONNECTION_STRING} = require("./constants");
const {migration} = require("./common");

// Function to transform document from db1 to the desired structure for db2
function transformDocument(doc) {
    return {
        _id: doc.id,
        name: doc.name?.trim(),
        surname: doc.surname?.trim(),
        email: doc.email?.trim(),
        phone: doc.phone?.trim(),
        note: doc.note?.trim(),
        version: doc.version,
        createdAt: doc.createdAt,
        updatedAt: doc.updatedAt,
        searchGrams: doc.searchGrams?.trim(),
    };
}

// Connection URL
const mongoConnectionString = MONGO_CONNECTION_STRING

// Database and collection names
const dbName1 = 'customers';
const dbName2 = 'customer-v2'
const db1CollectionName = 'customers'; // Replace with your actual collection name
const db2CollectionName = 'customers'; // Replace with your actual collection name

migration(
    mongoConnectionString,
    dbName1,
    dbName2,
    db1CollectionName,
    db2CollectionName,
    false,
    transformDocument
).then(r => console.log("Completed"));