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
        updatedAt: doc.updatedAt ?? new Date(),
        searchGrams: doc.searchGrams?.trim(),
    };
}

// Connection URL
const mongoConnectionString = MONGO_CONNECTION_STRING;

// Database and collection names
const dbName1 = 'customers';
const dbName2 = 'customer-v2'

migration(
    mongoConnectionString,
    dbName1,
    dbName2,
    "customers",
    "customers",
    false,
    transformDocument
).then(r => console.log("Completed Customer"));

migration(
    mongoConnectionString,
    dbName1,
    dbName2,
    "fidelitycards",
    "fidelitycards",
    false,
    transformFidelity
).then(r => console.log("Completed fidelitycards"));

migration(
    mongoConnectionString,
    dbName1,
    dbName2,
    "wallets",
    "wallets",
    false,
    transformWallet
).then(r => console.log("Completed wallets"));


function transformFidelity(doc) {
    return {
        _id: doc.id,
        customerId: doc.customerId,
        solariumPurchases: doc.solariumPurchases,
        vouchers: doc.vouchers.map(v => ({
            _id: v.id,
            _type: v._type,
            treatment: v.treatment,
            isUsed: v.isUsed,
            createdAt: v.createdAt
        })),
        createdAt: doc.createdAt,
        updatedAt: doc.updatedAt ?? new Date(),
    };
}

function transformWallet(doc) {
    return {
        _id: doc.id,
        owner: doc.owner,
        activeGiftCards: doc.activeGiftCards.map(g => ({
            _id: g.id,
            customerId: g.customerId,
            availableAmount: g.availableAmount,
            amount: g.amount,
            amountSpent: g.amountSpent,
            expireAt: g.expireAt,
            createdAt: g.createdAt,
        })),
        availableAmount: doc.availableAmount,
        spentAmount: doc.spentAmount,
        operations: doc.operations,
        createdAt: doc.createdAt,
        updatedAt: doc.updatedAt ?? new Date(),
    };
}