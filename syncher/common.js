const { MongoClient} = require('mongodb');

async function initialSync(collectionDb1, collectionDb2, transformFunction) {
    console.log("Starting initial sync...");
    let initialSyncCount = 0;
    const cursor = collectionDb1.find();
    while (await cursor.hasNext()) {
        const doc = await cursor.next();
        const transformedDoc = transformFunction(doc);
        if (initialSyncCount % 100 === 0) {
            console.log(`Sync ${initialSyncCount}`);
        }
        await collectionDb2.updateOne(
            { _id: transformedDoc._id },
            { $set: transformedDoc },
            { upsert: true }
        );
        initialSyncCount += 1;
    }
    console.log(`Initial sync completed. Total: ${initialSyncCount}`);
}

function startChangeStreamSync(collectionDb1, collectionDb2, transformFunction) {
    const changeStream = collectionDb1.watch({fullDocument: "updateLookup"});

    changeStream.on('change', async (change) => {
        const { operationType, fullDocument, documentKey } = change;

        if (operationType === 'insert' || operationType === 'replace' || operationType === 'update') {
            const transformedDoc = transformFunction(fullDocument);

            const updateResult = await collectionDb2.updateOne(
                { _id: transformedDoc._id },
                { $set: transformedDoc },
                { upsert: true }
            );

            if (updateResult.upsertedCount > 0) {
                console.log('Document inserted in db2');
            } else {
                console.log('Document updated in db2');
            }
        } else if (operationType === 'delete') {
            await collectionDb2.deleteOne({ _id: documentKey._id });
            console.log('Document deleted from db2');
        } else {
            console.warn(`Unknown operation type: ${operationType}`);
        }
    });

    console.log("Watching collection for changes...");
}

module.exports.migration = async function(
    mongoConnectionString,
    dbName1,
    dbName2,
    db1CollectionName,
    db2CollectionName,
    skipInitialSync,
    transformFunction
) {

    // Use connect method to connect to the server
    const clientDb1 = new MongoClient(mongoConnectionString, { useUnifiedTopology: true });
    const clientDb2 = new MongoClient(mongoConnectionString, { useUnifiedTopology: true });

    try {
        // Connect to both databases
        await clientDb1.connect();
        await clientDb2.connect();
        console.log("Connected successfully to both databases");

        const db1 = clientDb1.db(dbName1);
        const db2 = clientDb2.db(dbName2);
        const collectionDb1 = db1.collection(db1CollectionName);
        const collectionDb2 = db2.collection(db2CollectionName);

        // initial sync
        if (!skipInitialSync) {
            await initialSync(collectionDb1, collectionDb2, transformFunction);
        } else {
            console.log("Skipping initial sync...")
        }

        startChangeStreamSync(collectionDb1, collectionDb2, transformFunction);

    } catch (err) {
        console.error(err);
    }

}

