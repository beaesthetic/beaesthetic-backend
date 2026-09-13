# River UI

This deploys an internal River UI instance for production. It uses the
production `appointment-secrets-v2` Secret, so it displays only the River jobs
stored in the production appointment PostgreSQL database.

Access the production UI at https://river-ui.internal.k8s after the manifests
have synced.
The Service remains available for local troubleshooting with:

```bash
kubectl --context beaesthetic -n beaesthetic port-forward service/river-ui 8080:8080
```

Then open http://localhost:8080.

Before exposing it through an Ingress, add SSO or HTTP authentication. River
job arguments may contain customer data.

The development overlay is retained but intentionally excluded from the root
kustomization.

For a future service using a different PostgreSQL database, deploy a separate
River UI instance with that service's DSN. A single shared UI is possible only
when services deliberately use the same River database and schema.
