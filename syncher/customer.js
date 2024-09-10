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
const mongoConnectionString = "mongodb://localhost:27017/?authSource=admin&directConnection=true"; //MONGO_CONNECTION_STRING;

// Database and collection names
const dbName1 = 'customers';
const dbName2 = 'customer-v2'

// migration(
//     mongoConnectionString,
//     dbName1,
//     dbName2,
//     "customers",
//     "customers",
//     false,
//     transformDocument
// ).then(r => console.log("Completed Customer"));
//
// migration(
//     mongoConnectionString,
//     dbName1,
//     dbName2,
//     "fidelitycards",
//     "fidelitycards",
//     false,
//     transformFidelity
// ).then(r => console.log("Completed fidelitycards"));

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
            createdAt: v.createdAt ?? new Date(),
        })),
        createdAt: doc.createdAt,
        updatedAt: doc.updatedAt ?? new Date(),
    };
}

/*
{
  "_id": "b73f081a-950f-4b2e-8fc7-756566acda4d",
  "activeGiftCards": [
    {
      "_id": "3a6deb68-3dca-4516-881e-ecb22a10253b",
      "amount": 0,
      "amountSpent": 0,
      "availableAmount": 10,
      "createdAt": {
        "$date": "2024-07-26T09:56:59.743Z"
      },
      "customerId": "8ee3a426-02d6-4631-95ef-779edf303d52",
      "expireAt": {
        "$date": "2025-07-26T09:56:59.743Z"
      }
    }
  ],
  "availableAmount": 10,
  "createdAt": {
    "$date": "2024-07-26T09:56:59.743Z"
  },
  "operations": [
    {
      "type": "GiftCardMoneyCredited",
      "amount": 10,
      "at": {
        "$date": "2024-07-26T09:56:59.743Z"
      },
      "expireAt": {
        "$date": "2025-07-26T09:56:59.743Z"
      },
      "giftCardId": "3a6deb68-3dca-4516-881e-ecb22a10253b"
    }
  ],
  "owner": "8ee3a426-02d6-4631-95ef-779edf303d52",
  "spentAmount": 0,
  "updatedAt": {
    "$date": "2024-07-26T09:56:59.743Z"
  }
}
 */


function transformWallet(doc) {
    return {
        _id: doc.id,
        owner: doc.owner,
        activeGiftCards: doc.activeGiftCards.map(g => ({
            _id: g.id,
            amount: g.amount,
            amountSpent: g.amountSpent,
            availableAmount: g.availableAmount,
            createdAt: g.createdAt,
            customerId: g.customerId,
            expireAt: g.expireAt,
        })),
        availableAmount: doc.availableAmount,
        spentAmount: doc.spentAmount,
        operations: doc.operations,
        createdAt: doc.createdAt,
        updatedAt: doc.updatedAt ?? new Date(),
    };
}