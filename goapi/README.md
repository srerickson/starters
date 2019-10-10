# Go API Starter

*This is work in progress*


## Setup


Setup PostgreSQL server using Docker:
```bash
# set POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD as appropriate
docker run --rm -d --name myPostgres -p 5432:5432 -v $(pwd)/postgres_db:/var/lib/postgresql/data \
-e POSTGRES_DB=$USER -e POSTGRES_USER=$USER -e POSTGRES_PASSWORD=wordpass postgres:12-alpine
```

Connect to the server:
```bash
psql -h localhost -U $USER -d $USER
```

Create initial tables:
```sql
\i db/psql/schema.sql
```

#### Create self-signed TLS Certificate for development

```bash
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
```








