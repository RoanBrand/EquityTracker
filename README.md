# EquityTracker
A service that logs stock prices with a front-end for graphs and performance indicators.

### Install
- Install [Go](https://golang.org/) and set your `$GOPATH`.
- Install PostgreSQL 9.3+
- Create a database called "stocks" and run the following SQL:
```sql
CREATE TABLE "EOH"
(
rowid serial NOT NULL,
"TIMESTAMP" timestamp without time zone,
"ID" character varying,
"PRICE" double precision
)
WITH (
OIDS=FALSE
);
ALTER TABLE "EOH"
OWNER TO imqs;
```

### Use
- Open a terminal, then navigate to `$GOPATH/src/github.com/RoanBrand/EquityTracker`.
- Run `go run service.go`.
- Open `localhost` in your web browser.
