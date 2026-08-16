# River UI

This deploys one internal River UI instance for each environment. Each instance
uses that environment's `appointment-secrets-v2` Secret, so it displays only
the River jobs stored in the appointment PostgreSQL database.

Access the production UI at https://river-ui.internal.k8s and development at
https://river-ui-dev.internal.k8s after the manifests have synced.
The Service remains available for local troubleshooting with:

```bash
kubectl --context beaesthetic -n beaesthetic port-forward service/river-ui 8080:8080
```

Then open http://localhost:8080.

Before exposing it through an Ingress, add SSO or HTTP authentication. River
job arguments may contain customer data.

For a future service using a different PostgreSQL database, deploy a separate
River UI instance with that service's DSN. A single shared UI is possible only
when services deliberately use the same River database and schema.
