[
  {
    "$match": {
      "updatedAt": {
        "$gt": ISODate('{startDate}'),
        "$lte": ISODate('{endDate}')
      },
      "isCanceled": false,
      "migrated": null,
      "data.type": "appointment",
      "data.services": {
        "$type": "array"
      }
    }
  },
  {
    "$unwind": {
      "path": "$data.services"
    }
  },
  {
    "$project": {
      "serviceName": {
        "$trim": {
          "input": "$data.services"
        }
      },
      "date": "$start",
      "attendeeId": "$attendeeId"
    }
  },
  {
    "$group": {
      "_id": {
        "date": {
          "$concat": [
            "01-",
            {
              "$dateToString": {
                "date": "$date",
                "format": "%m-%Y"
              }
            }
          ]
        },
        "serviceName": "$serviceName"
      },
      "count": {
        "$sum": 1
      }
    }
  },
  {
    "$project": {
      "serviceName": "$_id.serviceName",
      "count": "$count",
      "date": {
        "$dateFromString": {
          "dateString": "$_id.date",
          "format": "%d-%m-%Y"
        }
      },
      "_id": 0
    }
  }
]