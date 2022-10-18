# etl-bitcoin

A repo for our Bitcoin ETL process.

The `client` package represents the connection to bitcoin node which will prob just be an extension of this https://github.com/toorop/go-bitcoind. The reason I defined our own interface at all is to define bulk operations if we're able to to do that with bitcoin rpc.

The `loader` package represents the link between the client and database. We may want to define different data to load in so I was thinking it would be configured in this package/step.

The `database` package just represents how to export data into some database format (csv, neo4j, sql, etc.).

For each of those packages we should define specific functions in the interfaces that specific implementations will need to define (like a neo4j implementation for database).
