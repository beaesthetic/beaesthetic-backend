const {MONGO_CONNECTION_STRING} = require("./constants");
const {migration} = require("./common");
const {MongoClient} = require("mongodb");

// Connection URL
const mongoConnectionString = MONGO_CONNECTION_STRING;

// Database and collection names
const dbName1 = 'appointment';
const dbName2 = 'appointments-v2'

async function main() {
    const clientDb1 = new MongoClient(mongoConnectionString, { useUnifiedTopology: true });
    await clientDb1.connect();
    const db1 = clientDb1.db(dbName1);
    const customers = db1.collection("customercontacts");
    const reminders = db1.collection("reminders");

    migration(
        mongoConnectionString,
        dbName1,
        dbName2,
        "agendaactivities",
        "agenda",
        false,
        doc => transformAgenda(doc, customers, reminders)
    ).then(r => console.log("Completed Agenda migrations"));

}

main()

/*
{
  "_id": "63f0d09ccc3f40d06a0eea58",
  "id": "009b9c41-0923-4099-b782-30cfcc5f6720",
  "start": "2022-09-30T07:00:00.000Z",
  "end": "2022-09-30T08:00:00.000Z",
  "attendeeId": "531b6ac6-a047-46be-8d13-0bfbcef4c5dc",
  "createdAt": "2022-09-09T06:36:22.000Z",
  "updatedAt": "2022-09-29T05:00:00.000Z",
  "isCanceled": false,
  "cancelReason": null,
  "reminderSent": true,
  "alarm": {
    "value": 1440,
    "unit": "MINUTES"
  },
  "data": {
    "type": "appointment",
    "services": [
      "Manicure smalto semipermanente"
    ]
  },
  "version": 4,
  "migrated": true
}*/

function createDisplayName(name, surname) {
    if (name?.trim()) {
        const surnameTrim = surname?.trim() ? ` ${surname}` : ""
        return `${name.trim()}${surnameTrim}`
    } else {
        return `${surname.trim()}`
    }
}

async function transformAgenda(doc, customers, reminders) {
    const customer = await customers.findOne({id: doc.attendeeId});
    const reminder = await reminders.findOne({appointmentId: doc.id})
    const displayName = createDisplayName(customer.name, customer.surname)
    let eventData;
    if (doc.data.type === "appointment") {
        eventData = {
            type: doc.data.type,
            services: doc.data.services
        }
    } else {
        eventData = {
            type: doc.data.type,
            title: doc.data.title,
            description: doc.data.description
        }
    }
    const reminderStatus = reminder?.sentAt ? "SENT" : doc.isCanceled ? "DELETED" : "PENDING"
    const cancelReason = doc.cancelReason ? doc.cancelReason : doc.isCanceled ? "NO_REASON" : null
    return {
        _id: doc.id,
        start: doc.start,
        end: doc.end,
        attendee: {
            id: doc.attendeeId,
            displayName: displayName
        },
        data: eventData,
        cancelReason: cancelReason,
        isCancelled: cancelReason != null ? true : doc.isCanceled,
        remindBeforeSeconds: (doc.alarm.value ?? 1440) * 60,
        reminderStatus: reminderStatus,
        createdAt: doc.createdAt ?? new Date(),
        updatedAt: doc.updatedAt ?? new Date(),
        version: doc.version
    };
}