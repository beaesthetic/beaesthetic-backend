# Appointment application flows

Questo documento descrive i flussi runtime correnti. Il nuovo percorso usa il modello `domain/v2`; le vecchie API OpenAPI e la coda scheduler restano temporaneamente attive come compatibilita' legacy.

## Runtime

Il comando `appointment app` avvia tramite il runtime condiviso:

- API HTTP legacy e API Calendar ProtoJSON;
- River client runtime con il worker `appointment.send_reminder`;
- consumer RabbitMQ dei lifecycle event;
- consumer della coda outcome `customer.notifications.outcomes`;
- consumer della vecchia scheduler queue, solo per drenare job gia' pubblicati.

River usa due client distinti:

- il client runtime registra e avvia i worker;
- il client producer e' usato da `RiverJobInserter` per `InsertTx`, `JobListTx` e `JobCancelTx` dentro la transazione presente nel `ContextDB`.

La separazione evita una dipendenza DI circolare tra worker, lifecycle service, scheduler e client River.

## Calendar create

Appointment gestisce un solo calendario. Se `calendarId` e' omesso viene applicato `d2a36e25-4824-4167-a062-a5af96f97703`; valori differenti non sono accettati.

Entry point:

```text
POST /v1/calendar-events
```

Sequenza:

1. L'adapter ProtoJSON converte la variante richiesta in un comando tipizzato: appointment, manual event o time block.
2. `CalendarService.Create` fa dispatch al subtype service corretto.
3. Per un appointment, il customer viene risolto prima di costruire l'aggregate.
4. La factory crea un unico `CalendarEvent` con il detail coerente con `event_type`.
5. Il dominio registra `CalendarEventCreated`.
6. Il repository apre una transazione e salva tabella base, detail, eventuali service item e reminder pending.
7. Gli eventi dominio vengono pubblicati nell'outbox nella stessa transazione.
8. Il consumer lifecycle ricarica la projection v2.
9. Solo se il detail e' `Appointment`, il lifecycle v2 pianifica il reminder e richiede la notifica di conferma.

Manual event e time block producono lo stesso lifecycle generico, ma il lifecycle appointment li ignora.

## Calendar update

Entry point:

```text
PATCH /v1/calendar-events/{id}
```

Sequenza:

1. L'adapter costruisce `UpdateEventCommand` con i soli campi modificabili.
2. `CalendarService.Update` carica l'aggregate e delega al subtype service.
3. `event_type` non e' modificabile.
4. Titolo e descrizione sono opzionali e appartengono solo ai detail che li prevedono.
5. Visibilita' e intervallo temporale appartengono al calendar event.
6. `CalendarEvent.Reschedule` emette `CalendarEventRescheduled` solo se start o end cambiano realmente.
7. Repository e outbox vengono salvati atomicamente.
8. Per un appointment rischedulato, il lifecycle cancella il vecchio job per key, inserisce il nuovo job e richiede `appointment_rescheduled` nella stessa transazione.

## Calendar cancel

Entry point:

```text
DELETE /v1/calendar-events/{id}
```

Sequenza:

1. Il servizio carica l'aggregate.
2. `CalendarEvent.Cancel` modella la cancellazione tramite `canceled_at` e reason, senza uno status generico.
3. Il dominio registra `CalendarEventCanceled`.
4. Repository e lifecycle outbox vengono salvati atomicamente.
5. Per gli appointment, il lifecycle cancella il job River identificato dalla key logica e marca il reminder `deleted`.

## Lifecycle dispatch

Il consumer accetta entrambi i formati durante il cutover:

- `CalendarEventCreated`, `CalendarEventRescheduled`, `CalendarEventCanceled` vengono gestiti da `AppointmentLifecycleService` v2;
- `AgendaEventScheduled`, `AgendaEventRescheduled`, `AgendaEventDeleted` vengono inoltrati al vecchio handler finche' la coda legacy non e' drenata.

L'evento lifecycle e' intenzionalmente generico. Il consumer successivo ricarica l'aggregate, osserva il detail e applica logica appointment solo quando necessaria.

## Reminder scheduling

`AppointmentLifecycleService` esegue in un'unica transazione:

1. carica `CalendarEventView`, composta dall'aggregate e dall'eventuale reminder;
2. calcola `sendAt` usando `remind_before`, soglia no-send e soglia immediate-send;
3. cancella gli eventuali job River non terminali con la stessa key;
4. inserisce il nuovo job con `InsertTx`;
5. aggiorna `appointment_reminders` a `scheduled`, oppure `unprocessable` se troppo tardi;
6. per create/reschedule pubblica la notification request su outbox;
7. salva il tracking in `appointment_notifications`.

La key applicativa e':

```text
appointment:{calendarEventID}:reminder
```

Appointment non salva l'ID tecnico del job River.

## Reminder execution

Quando il job `appointment.send_reminder` scade:

1. il worker chiama `AppointmentLifecycleService.SendDueReminder`;
2. il servizio apre una transazione e ricarica la projection v2;
3. termina senza errore se l'evento non e' un appointment, e' cancellato o il reminder non e' `scheduled`;
4. confronta `ExpectedStartAt` con lo start corrente per scartare job stale;
5. pubblica `appointment_reminder` su outbox;
6. salva `appointment_notifications` pending;
7. marca il reminder `send_requested`;
8. committa insieme job outcome applicativo, tracking e messaggio outbox.

Gli errori transitori fanno fallire il worker e consentono a River di applicare i retry configurati.

## Reminder resend

Entry point:

```text
POST /v1/calendar-events/{id}/reminder/resend
```

Sequenza:

1. valida event ID e idempotency key;
2. carica l'appointment e il reminder;
3. rifiuta eventi mancanti, cancellati, non-appointment o senza reminder;
4. pubblica la notification request e salva il tracking;
5. marca il reminder `send_requested`;
6. esegue tutto nella stessa transazione.

La key e':

```text
appointment:{calendarEventID}:reminder:resend:{requestKey}
```

## Notification outcome

Il consumer legge `CustomerNotificationOutcome` da `core-contracts/notification` e usa `notification_id` come correlation key.

`AppointmentLifecycleService.HandleNotificationOutcome` esegue atomicamente:

1. carica `appointment_notifications`;
2. marca la notifica `sent` oppure `failed`, conservando reason e message;
3. se il kind e' `reminder`, aggiorna anche `appointment_reminders` a `sent` o `failed`;
4. per confirmation e rescheduled non modifica il reminder.

Un outcome viene prodotto per ogni recipient/customer della request. Nel modello appointment corrente ogni richiesta ha un solo recipient customer.

## Customer notification request

`CustomerNotificationSender.SendCalendarNotification` costruisce il contratto condiviso e pubblica su outbox `customer.notifications`.

Tipi usati:

- `appointment_confirmation`;
- `appointment_rescheduled`;
- `appointment_reminder`.

Il servizio appointment non verifica preventivamente la presenza del numero di telefono. Notification decide se il recipient e' raggiungibile e pubblica l'outcome con failure reason, per esempio contatto assente.

## Legacy compatibility

Restano intenzionalmente temporanei:

- vecchie API OpenAPI basate su `domain.AgendaEvent`;
- `SchedulerQueueConsumer` e `ReminderSender` per messaggi gia' presenti nella vecchia coda;
- vecchi lifecycle event `AgendaEvent*`;
- colonne e query legacy necessarie al backfill e al periodo di stabilizzazione.

Il consumer outcome e il worker River non dipendono piu' da `domain.AgendaEvent`.

## Operational cutover

Usare la nuova immagine inizialmente come Job, senza avviare ancora `appointment app`.

1. Eseguire `appointment migrate up`, incluse le migration River.
2. Eseguire `appointment backfill-agenda-model` in dry-run e conservare il report.
3. Bloccare il cutover se `skipped_invalid_customers` e' maggiore di zero; questi appointment non possono essere ricostruiti con un customer UUID valido.
4. Eseguire `appointment schedule-future-reminders --dry-run` per stimare i job River da ricreare.
5. Fermare il vecchio deployment appointment o metterlo in maintenance, cosi' non scrive durante la migrazione.
6. Eseguire una sola volta `appointment backfill-agenda-model --execute`.
7. Eseguire `appointment schedule-future-reminders`; pianifica solo reminder legacy `PENDING` o `SCHEDULED`.
8. Avviare il nuovo deployment con `appointment app`.
9. Eseguire smoke test create, update senza cambio orario, reschedule, cancel, reminder e outcome.
10. Drenare la scheduler queue e i vecchi lifecycle event.
11. Rimuovere API, consumer, dominio e storage legacy in una migration successiva.

Il backfill normalizza gli eventi legacy `event` in `manual`. Non eseguirlo mentre il vecchio runtime sta ancora servendo traffico. Le notification legacy gia' scadute vengono riportate come `skipped_expired_notifications` e non vengono migrate come pending.
