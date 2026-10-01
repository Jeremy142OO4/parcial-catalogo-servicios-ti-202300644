# Backend

Backend Go del sistema. La organización prevista es:

```text
Handler/Controller -> Service -> Repository -> PostgreSQL
```

El comando `backend/cmd/importer` valida el Excel y puede persistirlo mediante `--database-url`. La persistencia se realiza en una transacción y conserva los nombres fuente, observaciones y referencias de origen.
